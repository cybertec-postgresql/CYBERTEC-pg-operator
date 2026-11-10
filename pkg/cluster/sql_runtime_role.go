package cluster

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const sqlRoleCheckTimeout = 10 * time.Second

func physicalSQLStandby(db *sql.DB) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sqlRoleCheckTimeout)
	defer cancel()

	var recovery bool
	if err := db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&recovery); err != nil {
		return false, fmt.Errorf("could not verify SQL recovery state: %w", err)
	}
	return recovery, nil
}

// sqlWritesAllowed checks the current database role on every reconciliation.
// A coordinator can configure physical recovery without spec.standbyCluster.
func (c *Cluster) sqlWritesAllowed() (bool, error) {
	if err := c.initDbConn(); err != nil {
		return false, fmt.Errorf("could not connect for SQL role verification: %w", err)
	}
	defer c.closeDbConn()

	standby, err := physicalSQLStandby(c.pgDb)
	if err != nil {
		return false, err
	}
	if standby {
		c.logger.Info("physical SQL standby detected; skipping role, database and schema writes")
	}
	return !standby, nil
}
