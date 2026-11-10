package cluster

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
)

type recoveryConnector struct {
	conn *recoveryConn
}

func (connector recoveryConnector) Connect(context.Context) (driver.Conn, error) {
	return connector.conn, nil
}

func (recoveryConnector) Driver() driver.Driver {
	return recoveryDriver{}
}

type recoveryDriver struct{}

func (recoveryDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type recoveryConn struct {
	values    []driver.Value
	queryErr  error
	execErr   error
	calls     int
	execCalls int
	closed    bool
}

func (*recoveryConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}

func (conn *recoveryConn) Close() error {
	conn.closed = true
	return nil
}

func (*recoveryConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}

func (conn *recoveryConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if _, ok := ctx.Deadline(); !ok {
		return nil, errors.New("missing query deadline")
	}
	if query != "SELECT pg_is_in_recovery()" || len(args) != 0 {
		return nil, errors.New("unexpected SQL")
	}
	conn.calls++
	if conn.queryErr != nil {
		return nil, conn.queryErr
	}
	return &recoveryRows{values: conn.values}, nil
}

func (conn *recoveryConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if query != cpoMonitoring || len(args) != 0 {
		return nil, errors.New("unexpected SQL")
	}
	conn.execCalls++
	if conn.execErr != nil {
		return nil, conn.execErr
	}
	return driver.RowsAffected(1), nil
}

type recoveryRows struct {
	values []driver.Value
	index  int
}

func (*recoveryRows) Columns() []string {
	return []string{"pg_is_in_recovery"}
}

func (*recoveryRows) Close() error {
	return nil
}

func (rows *recoveryRows) Next(destination []driver.Value) error {
	if rows.index >= len(rows.values) {
		return io.EOF
	}
	destination[0] = rows.values[rows.index]
	rows.index++
	return nil
}

func TestPhysicalSQLStandbyFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name     string
		values   []driver.Value
		queryErr error
		standby  bool
		wantErr  bool
	}{
		{name: "primary", values: []driver.Value{false}},
		{name: "standby", values: []driver.Value{true}, standby: true},
		{name: "query error", queryErr: errors.New("query failed"), wantErr: true},
		{name: "timeout", queryErr: context.DeadlineExceeded, wantErr: true},
		{name: "empty", wantErr: true},
		{name: "null", values: []driver.Value{nil}, wantErr: true},
		{name: "invalid", values: []driver.Value{"not-a-boolean"}, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			conn := &recoveryConn{values: test.values, queryErr: test.queryErr}
			db := sql.OpenDB(recoveryConnector{conn: conn})
			defer db.Close()

			standby, err := physicalSQLStandby(db)
			if standby != test.standby || (err != nil) != test.wantErr {
				t.Fatalf("got standby=%v error=%v", standby, err)
			}
			if conn.calls != 1 {
				t.Fatalf("expected one role query, got %d", conn.calls)
			}
		})
	}
}

func TestPhysicalSQLStandbyRechecksAfterPromotion(t *testing.T) {
	conn := &recoveryConn{values: []driver.Value{true}}
	db := sql.OpenDB(recoveryConnector{conn: conn})
	defer db.Close()

	if standby, err := physicalSQLStandby(db); !standby || err != nil {
		t.Fatalf("standby was not detected: %v", err)
	}
	conn.values = []driver.Value{false}
	if standby, err := physicalSQLStandby(db); standby || err != nil {
		t.Fatalf("promotion used a cached role: %v", err)
	}
	if conn.calls != 2 {
		t.Fatalf("expected two role queries, got %d", conn.calls)
	}
}

func TestReconcileMonitoringPermissionsChecksRuntimeRole(t *testing.T) {
	execErr := errors.New("monitoring DDL failed")
	queryErr := errors.New("role query failed")

	for _, test := range []struct {
		name      string
		conn      *recoveryConn
		written   bool
		wantExec  int
		wantError error
	}{
		{
			name:     "primary writes",
			conn:     &recoveryConn{values: []driver.Value{false}},
			written:  true,
			wantExec: 1,
		},
		{
			name: "standby skips",
			conn: &recoveryConn{values: []driver.Value{true}},
		},
		{
			name:      "role query error fails closed",
			conn:      &recoveryConn{queryErr: queryErr},
			wantError: queryErr,
		},
		{
			name:      "DDL error is returned",
			conn:      &recoveryConn{values: []driver.Value{false}, execErr: execErr},
			wantExec:  1,
			wantError: execErr,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := sql.OpenDB(recoveryConnector{conn: test.conn})
			defer db.Close()

			written, err := reconcileMonitoringPermissions(db)
			if written != test.written {
				t.Fatalf("written=%v, want %v", written, test.written)
			}
			if test.wantError == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantError != nil && !errors.Is(err, test.wantError) {
				t.Fatalf("error=%v, want %v", err, test.wantError)
			}
			if test.conn.calls != 1 {
				t.Fatalf("role queries=%d, want 1", test.conn.calls)
			}
			if test.conn.execCalls != test.wantExec {
				t.Fatalf("DDL executions=%d, want %d", test.conn.execCalls, test.wantExec)
			}
		})
	}
}

func TestSQLWritesAllowedClosesConnection(t *testing.T) {
	for _, standby := range []bool{false, true} {
		conn := &recoveryConn{values: []driver.Value{standby}}
		db := sql.OpenDB(recoveryConnector{conn: conn})
		cluster := &Cluster{
			pgDb:   db,
			logger: logrus.New().WithField("test", "sql-runtime-role"),
		}

		allowed, err := cluster.sqlWritesAllowed()
		if err != nil || allowed == standby {
			t.Fatalf("standby=%v allowed=%v error=%v", standby, allowed, err)
		}
		if cluster.pgDb != nil || !conn.closed {
			t.Fatal("verification connection was not closed")
		}
	}
}

func TestSQLWritesAllowedReturnsQueryError(t *testing.T) {
	queryErr := errors.New("verification failed")
	conn := &recoveryConn{queryErr: queryErr}
	cluster := &Cluster{
		pgDb:   sql.OpenDB(recoveryConnector{conn: conn}),
		logger: logrus.New().WithField("test", "sql-runtime-role"),
	}

	allowed, err := cluster.sqlWritesAllowed()
	if allowed || !errors.Is(err, queryErr) {
		t.Fatalf("error was hidden: allowed=%v error=%v", allowed, err)
	}
	if cluster.pgDb != nil || !conn.closed {
		t.Fatal("error-path connection was not closed")
	}
}
