package types

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Database defines what methods are needed to be implemented by a database
type Database interface {
	Open(connStr, appName, minMigration string, maxLifeTime, maxOpenConn, mapIdleConn, logLevel int) (err error)
	NewTx(ctx context.Context, readOnly bool) (tx Transaction, err error)
	Close()
}

// Transaction defines what methods are needed to be implmented by a transaction
type Transaction interface {
	Commit() (err error)
	Rollback()

	Query(query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(query string, args ...interface{}) pgx.Row
	Exec(query string, args ...interface{}) (pgconn.CommandTag, error)
}
