package cluster

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
)

type closeErrorConnector struct {
	conn *closeErrorConn
}

func (connector closeErrorConnector) Connect(context.Context) (driver.Conn, error) {
	return connector.conn, nil
}

func (closeErrorConnector) Driver() driver.Driver {
	return closeErrorDriver{}
}

type closeErrorDriver struct{}

func (closeErrorDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use connector")
}

type closeErrorConn struct {
	err    error
	closed bool
}

func (*closeErrorConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}

func (conn *closeErrorConn) Close() error {
	conn.closed = true
	return conn.err
}

func (*closeErrorConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}

func TestCloseDbConnReturnsDriverError(t *testing.T) {
	closeErr := errors.New("close failed")
	conn := &closeErrorConn{err: closeErr}
	db := sql.OpenDB(closeErrorConnector{conn: conn})
	if err := db.Ping(); err != nil {
		t.Fatalf("could not open test connection: %v", err)
	}

	cluster := &Cluster{
		pgDb:   db,
		logger: logrus.New().WithField("test", "database-close"),
	}

	if err := cluster.closeDbConn(); !errors.Is(err, closeErr) {
		t.Fatalf("close error was hidden: %v", err)
	}
	if cluster.pgDb != nil || !conn.closed {
		t.Fatal("database connection was not cleared")
	}
}
