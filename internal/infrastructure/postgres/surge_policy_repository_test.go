package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

var policyRepoDriverID atomic.Uint64

type policyRepoDriver struct{ rows [][]driver.Value }
type policyRepoConn struct{ rows [][]driver.Value }
type policyRepoRows struct {
	rows  [][]driver.Value
	index int
}

func (d policyRepoDriver) Open(string) (driver.Conn, error) { return policyRepoConn{rows: d.rows}, nil }
func (policyRepoConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported")
}
func (policyRepoConn) Close() error { return nil }
func (policyRepoConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions not supported")
}
func (c policyRepoConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &policyRepoRows{rows: c.rows}, nil
}
func (*policyRepoRows) Columns() []string {
	return []string{"id", "locality_id", "version", "status", "min_multiplier", "max_multiplier", "minimum_demand", "effective_from", "effective_until", "min_imbalance_ratio", "multiplier"}
}
func (*policyRepoRows) Close() error { return nil }
func (r *policyRepoRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func TestSurgePolicyRepositoryLoadsOrderedTiers(t *testing.T) {
	id, localityID := uuid.New(), uuid.New()
	effectiveFrom := time.Now().Add(-time.Hour)
	stateRows := [][]driver.Value{
		{id.String(), localityID.String(), int64(1), "ACTIVE", float64(1), float64(1.5), int64(3), effectiveFrom, nil, float64(1), float64(1.1)},
		{id.String(), localityID.String(), int64(1), "ACTIVE", float64(1), float64(1.5), int64(3), effectiveFrom, nil, float64(2), float64(1.5)},
	}
	name := fmt.Sprintf("surge-policy-repository-test-%d", policyRepoDriverID.Add(1))
	sql.Register(name, policyRepoDriver{rows: stateRows})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	policies, err := NewSurgePolicyRepository(db).LoadActive(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) != 1 || len(policies[0].RatioTiers) != 2 {
		t.Fatalf("loaded policies = %+v", policies)
	}
	if policies[0].MinimumDemand != 3 || policies[0].RatioTiers[0].MinImbalanceRatio != 1 || policies[0].RatioTiers[1].Multiplier != 1.5 {
		t.Fatalf("loaded policy fields or tier order are wrong: %+v", policies[0])
	}
}
