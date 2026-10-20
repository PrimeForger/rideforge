package application

import (
	"context"
	"testing"
	"time"
)

type refreshFunc func(context.Context) error

func (f refreshFunc) Refresh(ctx context.Context) error { return f(ctx) }

type surgeObserverStub struct {
	decisions []string
	refreshes []string
}

func (o *surgeObserverStub) ObserveSurgeDecision(result string) {
	o.decisions = append(o.decisions, result)
}
func (o *surgeObserverStub) ObserveSurgePolicyRefresh(result string) {
	o.refreshes = append(o.refreshes, result)
}

func TestSurgePolicyRefreshWorkerRefreshesImmediatelyAndStopsOnCancellation(t *testing.T) {
	called := make(chan struct{}, 1)
	observer := &surgeObserverStub{}
	worker := NewSurgePolicyRefreshWorker(refreshFunc(func(context.Context) error {
		called <- struct{}{}
		return nil
	}), time.Minute, observer, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Start(ctx)
		close(done)
	}()
	select {
	case <-called:
	case <-time.After(time.Second):
		cancel()
		t.Fatal("worker did not refresh immediately")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
	if len(observer.refreshes) != 1 || observer.refreshes[0] != "success" {
		t.Fatalf("refresh observations = %v", observer.refreshes)
	}
}

func TestSurgePolicyRefreshWorkerRecordsFailure(t *testing.T) {
	observer := &surgeObserverStub{}
	worker := NewSurgePolicyRefreshWorker(refreshFunc(func(context.Context) error { return context.DeadlineExceeded }), time.Minute, observer, nil)
	worker.refresh(context.Background())
	if len(observer.refreshes) != 1 || observer.refreshes[0] != "error" {
		t.Fatalf("refresh observations = %v", observer.refreshes)
	}
}
