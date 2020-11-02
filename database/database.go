package database

import (
	"context"
	"lab/iam/config"
	"lab/iam/database/postgres"
	"lab/iam/database/types"
)

var (
	roDB types.Database
	rwDB types.Database
)

// OpenDatabases tries to open connection with all set-up databases
func OpenDatabases() (err error) {
	var (
		cfg = config.GetConfig()
	)

	roDB, rwDB = postgres.New(), postgres.New()

	if err = roDB.Open(
		cfg.Database.RODatabase,
		cfg.ApplicationName,
		cfg.Database.MinimunMigration,
		cfg.Database.MaxConnLifetime,
		cfg.Database.MaxOpenConn,
		cfg.Database.MaxIdleConn,
		cfg.LogLevel,
	); err != nil {
		return
	}

	if err = rwDB.Open(
		cfg.Database.RWDatabase,
		cfg.ApplicationName,
		cfg.Database.MinimunMigration,
		cfg.Database.MaxConnLifetime,
		cfg.Database.MaxOpenConn,
		cfg.Database.MaxIdleConn,
		cfg.LogLevel,
	); err != nil {
		return
	}

	return
}

// Close tries to close all dangling connections
func Close() {
	roDB.Close()
	rwDB.Close()
}

// NewTx returns a new database transaction
func NewTx(ctx context.Context, readOnly bool) (tx types.Transaction, err error) {
	var (
		db types.Database = rwDB
	)

	if readOnly {
		db = roDB
	}

	return db.NewTx(ctx, readOnly)
}
