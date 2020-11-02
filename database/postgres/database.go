package postgres

import (
	"context"
	"lab/iam/database/types"
	"lab/iam/utils"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/washingt0/oops"
	"go.uber.org/zap"
)

type pg struct {
	db *pgxpool.Pool
}

// New returns a new instance of types.Database
func New() types.Database {
	return &pg{}
}

// Open tries to connect with the given postgres server
func (p *pg) Open(connStr, appName, minMigration string, maxLifeTime, maxOpenConn, mapIdleConn, logLevel int) (err error) {
	var (
		poolCfg *pgxpool.Config
	)

	if poolCfg, err = pgxpool.ParseConfig(connStr); err != nil {
		return oops.ThrowError("invalid connStr", err)
	}

	poolCfg.MaxConns = int32(maxOpenConn)
	poolCfg.MaxConnLifetime = time.Duration(maxLifeTime) * time.Second
	poolCfg.ConnConfig.ConnectTimeout = 3 * time.Second
	poolCfg.ConnConfig.Logger = zapadapter.NewLogger(zap.L())
	poolCfg.ConnConfig.LogLevel = pgx.LogLevel(6 - logLevel)
	poolCfg.ConnConfig.RuntimeParams["search_path"] = "public"
	poolCfg.ConnConfig.RuntimeParams["application_name"] = appName
	poolCfg.ConnConfig.RuntimeParams["DateStyle"] = "ISO"
	poolCfg.ConnConfig.RuntimeParams["IntervalStyle"] = "iso_8601"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if p.db, err = pgxpool.ConnectConfig(
		ctx,
		poolCfg,
	); err != nil {
		oops.ThrowError("unable to connect with database server", err)
	}

	var validMigration bool
	if err = p.db.QueryRow(ctx, `
		SELECT COUNT(1) > 0
		FROM public.t_migration
		WHERE name = $1
		  AND rolled_back = FALSE
	`, minMigration).Scan(&validMigration); err != nil {
		return oops.ThrowError("error when checking migration version", err)
	}

	if !validMigration {
		return oops.ThrowError("invalid migration version", nil)
	}

	return
}

// NewTx creates a new transaction with the current database
func (p *pg) NewTx(ctx context.Context, readOnly bool) (types.Transaction, error) {
	var (
		dbTx *tx = &tx{
			ctx: ctx,
		}
		err error
	)

	ctx2, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	var accMode pgx.TxAccessMode = pgx.ReadWrite
	if readOnly {
		accMode = pgx.ReadOnly
	}

	if dbTx.tx, err = p.db.BeginTx(ctx2, pgx.TxOptions{
		AccessMode:     accMode,
		IsoLevel:       pgx.Serializable,
		DeferrableMode: pgx.Deferrable,
	}); err != nil {
		return nil, err
	}

	if err = prepareTx(ctx, dbTx.tx); err != nil {
		return nil, err
	}

	return dbTx, nil
}

// Close closes the current connection pool
func (p *pg) Close() {
	p.db.Close()
}

type tx struct {
	ctx context.Context
	tx  pgx.Tx
}

// Commit persists all changes for the current transaction
func (t *tx) Commit() (err error) {
	return t.tx.Commit(t.ctx)
}

// Rollback revertes all changes for the current transaction
func (t *tx) Rollback() {
	_ = t.tx.Rollback(t.ctx)
}

// Query fetches from the database a set of resultant rows
func (t *tx) Query(query string, args ...interface{}) (rows pgx.Rows, err error) {
	return t.tx.Query(t.ctx, getCaller(t.ctx)+query, args...)
}

// QueryRow fetches a single row
func (t *tx) QueryRow(query string, args ...interface{}) (row pgx.Row) {
	return t.tx.QueryRow(t.ctx, getCaller(t.ctx)+query, args...)
}

// Exec performs a operation that not returns a row
func (t *tx) Exec(query string, args ...interface{}) (out pgconn.CommandTag, err error) {
	return t.tx.Exec(t.ctx, getCaller(t.ctx)+query, args...)
}

func getCaller(ctx context.Context) (out string) {
	var (
		file      string
		line      int
		requestID *string
	)

	if ctx.Value("RID") != nil {
		requestID = utils.GetStringPointer(ctx.Value("RID").(string))
	}

	if requestID != nil {
		out = "-- Request: " + *requestID + " \r\n"
	}

	for i := 1; i < 8; i++ {
		_, file, line, _ = runtime.Caller(i)
		if strings.Contains(file, "iam/") {
			out += "-- " + strconv.FormatInt(int64(i), 10) + " :: " + file + ":" + strconv.FormatInt(int64(line), 10) + "\r\n"
		}
	}

	return
}

func prepareTx(ctx context.Context, tx pgx.Tx) (err error) {
	var (
		requestID *string
		userID    *string
	)

	if ctx.Value("RID") != nil {
		requestID = utils.GetStringPointer(ctx.Value("RID").(string))

		if _, err = tx.Exec(ctx, "SET LOCAL application.request_id TO '"+*requestID+"'"); err != nil {
			return oops.ThrowError("unable to set requestID", err)
		}
	}

	if ctx.Value("UID") != nil {
		userID = utils.GetStringPointer(ctx.Value("UID").(string))

		if _, err = tx.Exec(ctx, "SET LOCAL application.user_id TO '"+*userID+"'"); err != nil {
			return oops.ThrowError("unable to set userID", err)
		}
	}

	return
}
