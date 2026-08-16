package application

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

type fakeRecoveryRunner struct {
	mu           sync.Mutex
	calls        int
	blockCh      chan struct{}
	concurrent   int32
	maxConcur    int32
	returnErr    error
	returnResult RecoveryResult
}

func (f *fakeRecoveryRunner) RebuildIndex(ctx context.Context) (RecoveryResult, error) {
	cur := atomic.AddInt32(&f.concurrent, 1)
	for {
		max := atomic.LoadInt32(&f.maxConcur)
		if cur <= max {
			break
		}
		if atomic.CompareAndSwapInt32(&f.maxConcur, max, cur) {
			break
		}
	}
	defer atomic.AddInt32(&f.concurrent, -1)

	if f.blockCh != nil {
		select {
		case <-f.blockCh:
		case <-ctx.Done():
			return RecoveryResult{}, ctx.Err()
		}
	}

	f.mu.Lock()
	f.calls++
	err := f.returnErr
	res := f.returnResult
	f.mu.Unlock()

	return res, err
}

func (f *fakeRecoveryRunner) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func waitForCalls(runner *fakeRecoveryRunner, target int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if runner.CallCount() >= target {
			return true
		}
		time.Sleep(2 * time.Millisecond)
	}
	return runner.CallCount() >= target
}

func TestH3ReconciliationWorker_Disabled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	runner := &fakeRecoveryRunner{}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, false, 10*time.Millisecond, log)

	worker.Start(ctx) // Should return immediately because disabled

	if runner.CallCount() != 0 {
		t.Errorf("expected 0 calls for disabled worker, got %d", runner.CallCount())
	}
}

func TestH3ReconciliationWorker_PeriodicExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &fakeRecoveryRunner{
		returnResult: RecoveryResult{TotalDrivers: 5, RestoredDrivers: 5},
	}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, true, 10*time.Millisecond, log)

	go worker.Start(ctx)

	if !waitForCalls(runner, 2, 200*time.Millisecond) {
		t.Errorf("expected at least 2 periodic reconciliation runs, got %d", runner.CallCount())
	}
}

func TestH3ReconciliationWorker_FailureResilience(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &fakeRecoveryRunner{
		returnErr: errors.New("simulated recovery failure"),
	}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, true, 10*time.Millisecond, log)

	go worker.Start(ctx)

	// Worker should survive failures and continue ticks
	if !waitForCalls(runner, 2, 200*time.Millisecond) {
		t.Errorf("expected worker to survive failure and continue periodic runs, got %d calls", runner.CallCount())
	}
}

func TestH3ReconciliationWorker_GracefulCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &fakeRecoveryRunner{}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, true, 1*time.Second, log)

	doneCh := make(chan struct{})
	go func() {
		worker.Start(ctx)
		close(doneCh)
	}()

	cancel()

	select {
	case <-doneCh:
		// Exited cleanly
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not exit cleanly upon context cancellation")
	}
}

func TestH3ReconciliationWorker_NoOverlappingExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockCh := make(chan struct{})
	runner := &fakeRecoveryRunner{
		blockCh: blockCh,
	}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, true, 1*time.Hour, log)

	// Trigger run 1 in background (will block on blockCh)
	go worker.runReconciliation(ctx)

	// Poll until run 1 is actively executing inside blockCh
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) && atomic.LoadInt32(&runner.concurrent) == 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Trigger run 2 while run 1 is active
	worker.runReconciliation(ctx) // Should be skipped immediately

	// Unblock run 1
	close(blockCh)

	// Wait for run 1 to finish
	waitForCalls(runner, 1, 100*time.Millisecond)

	if runner.CallCount() != 1 {
		t.Errorf("expected exactly 1 completed run (2nd run skipped), got %d", runner.CallCount())
	}

	if atomic.LoadInt32(&runner.maxConcur) > 1 {
		t.Errorf("expected max concurrent runs <= 1, got %d", runner.maxConcur)
	}
}

func TestH3ReconciliationWorker_SuccessfulResult(t *testing.T) {
	ctx := context.Background()
	expectedResult := RecoveryResult{
		TotalDrivers:    10,
		RestoredDrivers: 8,
		SkippedDrivers:  2,
		CellUpdates:     3,
		Duration:        12 * time.Millisecond,
	}
	runner := &fakeRecoveryRunner{
		returnResult: expectedResult,
	}
	log := zap.NewNop()

	worker := NewH3ReconciliationWorker(runner, true, 1*time.Hour, log)

	worker.runReconciliation(ctx)

	if runner.CallCount() != 1 {
		t.Errorf("expected 1 call, got %d", runner.CallCount())
	}
}
