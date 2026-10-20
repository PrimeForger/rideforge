package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/google/uuid"
)

var localityRepoDriverID atomic.Uint64

type localityRepoDBState struct {
	queries        []string
	failCellInsert bool
	committed      bool
	rolledBack     bool
}

type localityRepoDriver struct{ state *localityRepoDBState }
type localityRepoConn struct{ state *localityRepoDBState }
type localityRepoTx struct{ state *localityRepoDBState }

func (d localityRepoDriver) Open(string) (driver.Conn, error) {
	return localityRepoConn{state: d.state}, nil
}
func (c localityRepoConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}
func (c localityRepoConn) Close() error              { return nil }
func (c localityRepoConn) Begin() (driver.Tx, error) { return localityRepoTx{state: c.state}, nil }
func (c localityRepoConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return localityRepoTx{state: c.state}, nil
}
func (c localityRepoConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.queries = append(c.state.queries, query)
	if c.state.failCellInsert && strings.Contains(query, "INSERT INTO locality_h3_cells") {
		return nil, errors.New("duplicate H3 owner")
	}
	return driver.RowsAffected(1), nil
}
func (tx localityRepoTx) Commit() error {
	tx.state.committed = true
	return nil
}
func (tx localityRepoTx) Rollback() error {
	tx.state.rolledBack = true
	return nil
}

var _ driver.Driver = localityRepoDriver{}
var _ driver.ConnBeginTx = localityRepoConn{}
var _ driver.ExecerContext = localityRepoConn{}

func newLocalityRepositoryTestDB(t *testing.T, state *localityRepoDBState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("locality-repository-test-%d", localityRepoDriverID.Add(1))
	sql.Register(name, localityRepoDriver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestLocalityBoundaryAndH3CoverageCommitTogether(t *testing.T) {
	state := &localityRepoDBState{}
	repo := NewLocalityRepository(newLocalityRepositoryTestDB(t, state), geo.NewH3Service(8, 1))
	if err := repo.UpdateBoundary(context.Background(), uuid.New(), "MULTIPOLYGON EMPTY", []string{"cell-a"}); err != nil {
		t.Fatal(err)
	}
	if !state.committed || state.rolledBack {
		t.Fatalf("transaction outcome committed=%t rolledBack=%t", state.committed, state.rolledBack)
	}
	if len(state.queries) != 3 || !strings.Contains(state.queries[0], "UPDATE localities") || !strings.Contains(state.queries[1], "DELETE FROM locality_h3_cells") || !strings.Contains(state.queries[2], "INSERT INTO locality_h3_cells") {
		t.Fatalf("unexpected update sequence: %v", state.queries)
	}
}

func TestLocalityCoverageInsertFailureRollsBackBoundaryUpdate(t *testing.T) {
	state := &localityRepoDBState{failCellInsert: true}
	repo := NewLocalityRepository(newLocalityRepositoryTestDB(t, state), geo.NewH3Service(8, 1))
	err := repo.UpdateBoundary(context.Background(), uuid.New(), "MULTIPOLYGON EMPTY", []string{"owned-cell"})
	if err == nil {
		t.Fatal("expected H3 ownership conflict")
	}
	if state.committed || !state.rolledBack {
		t.Fatalf("partial boundary update was not rolled back: committed=%t rolledBack=%t", state.committed, state.rolledBack)
	}
}
