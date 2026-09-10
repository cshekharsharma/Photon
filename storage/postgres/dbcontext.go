package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBContext defines a contextual wrapper around PostgreSQL execution logic.
// It allows SQL operations to be executed using:
//
//  1. An existing transaction (Tx)
//  2. A manually supplied database connection (Conn)
//  3. A lazily resolved cluster-based connection (Cluster)
//
// DBContext also supports function overrides (ExecFn, PrepareFn, QueryFn)
// to inject custom behavior such as:
//   - Unit test mocks
//   - Observability / tracing wrappers
//   - Query auditing
//   - Fault injection
//
// IMPORTANT: Prefer the Context variants (ExecContext/QueryContext/PrepareContext).
// The non-context methods call the context variants with context.Background().
type DBContext struct {
	// Tx represents an active PostgreSQL transaction.
	// If set, all SQL operations are executed inside this transaction.
	Tx Tx

	// Conn is a direct PostgreSQL connection interface.
	// Used when Tx is nil.
	Conn PostgresDbInterface

	// Cluster is a named PostgreSQL cluster.
	// Used to lazily obtain a connection if both Tx and Conn are nil.
	Cluster string

	// DefaultTimeout is applied when callers use non-context methods
	// (Exec/Query/Prepare) or pass a context without deadline.
	//
	// If <= 0, no implicit deadline is added.
	DefaultTimeout time.Duration

	// ExecFn optionally overrides execution of non-query statements.
	ExecFn func(ctx context.Context, query string, args ...any) (sql.Result, error)

	// PrepareFn optionally overrides preparation of SQL statements.
	PrepareFn func(ctx context.Context, query string) (*sql.Stmt, error)

	// QueryFn optionally overrides execution of queries returning rows.
	QueryFn func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Exec executes a SQL statement that does not return rows.
// It uses context.Background() and applies DefaultTimeout if configured.
func (ctx *DBContext) Exec(query string, args ...any) (sql.Result, error) {
	return ctx.ExecContext(context.Background(), query, args...)
}

// Query executes a SQL query that returns rows.
// It uses context.Background() and applies DefaultTimeout if configured.
func (ctx *DBContext) Query(query string, args ...any) (*sql.Rows, error) {
	return ctx.QueryContext(context.Background(), query, args...)
}

// Prepare creates a prepared SQL statement.
// It uses context.Background() and applies DefaultTimeout if configured.
func (ctx *DBContext) Prepare(query string) (*sql.Stmt, error) {
	return ctx.PrepareContext(context.Background(), query)
}

// ExecContext executes a SQL statement that does not return rows
// (INSERT, UPDATE, DELETE, etc).
//
// Resolution order:
//  1. ExecFn override (if set)
//  2. Tx.ExecContext() / Tx.Exec()
//  3. Conn.ExecContext() / Conn.Exec()
//  4. Cluster-based connection via Connect()
func (ctx *DBContext) ExecContext(c context.Context, query string, args ...any) (sql.Result, error) {
	c, cancel := ctx.withDefaultTimeout(c)
	if cancel != nil {
		defer cancel()
	}

	if ctx.ExecFn != nil {
		// TODO(otel): start span "postgres.exec" with query metadata (sanitized), args count, cluster, etc.
		return ctx.ExecFn(c, query, args...)
	}

	return ctx.execContext(c, query, args...)
}

// PrepareContext prepares a SQL statement.
//
// Resolution order:
//  1. PrepareFn override (if set)
//  2. Tx.PrepareContext() / Tx.Prepare()
//  3. Conn.PrepareContext() / Conn.Prepare()
//  4. Cluster-based connection via Connect()
func (ctx *DBContext) PrepareContext(c context.Context, query string) (*sql.Stmt, error) {
	c, cancel := ctx.withDefaultTimeout(c)
	if cancel != nil {
		defer cancel()
	}

	if ctx.PrepareFn != nil {
		// TODO(otel): start span "postgres.prepare"
		return ctx.PrepareFn(c, query)
	}

	return ctx.prepareContext(c, query)
}

// QueryContext executes a SQL query that returns rows.
//
// Resolution order:
//  1. QueryFn override (if set)
//  2. Tx.QueryContext() / Tx.Query()
//  3. Conn.QueryContext() / Conn.Query()
//  4. Cluster-based connection via Connect()
func (ctx *DBContext) QueryContext(c context.Context, query string, args ...any) (*sql.Rows, error) {
	c, cancel := ctx.withDefaultTimeout(c)
	if cancel != nil {
		defer cancel()
	}

	if ctx.QueryFn != nil {
		// TODO(otel): start span "postgres.query"
		return ctx.QueryFn(c, query, args...)
	}

	return ctx.queryContext(c, query, args...)
}

// --- internals ---

func (ctx *DBContext) withDefaultTimeout(c context.Context) (context.Context, context.CancelFunc) {
	if c == nil {
		c = context.Background()
	}
	if ctx.DefaultTimeout <= 0 {
		return c, nil
	}
	if _, ok := c.Deadline(); ok {
		return c, nil
	}
	return context.WithTimeout(c, ctx.DefaultTimeout)
}

func (ctx *DBContext) validateResolution() error {
	if ctx.Tx == nil && ctx.Conn == nil && ctx.Cluster == "" {
		return fmt.Errorf("postgres: DBContext has no Tx, no Conn, and empty Cluster")
	}
	return nil
}

func (ctx *DBContext) resolveConn() (PostgresDbInterface, error) {
	if err := ctx.validateResolution(); err != nil {
		return nil, err
	}

	if ctx.Tx != nil || ctx.Conn != nil {
		// caller won’t reach here normally, but keep clean behavior
		return nil, fmt.Errorf("postgres: resolveConn called even though Tx/Conn present")
	}

	// TODO(otel): record cluster connection lookup
	return Connect(helperPostgresConnector, ctx.Cluster)
}

func (ctx *DBContext) execContext(c context.Context, query string, args ...any) (sql.Result, error) {
	if err := ctx.validateResolution(); err != nil {
		return nil, err
	}

	if ctx.Tx != nil {
		if t, ok := ctx.Tx.(TxContext); ok {
			return t.ExecContext(c, query, args...)
		}
		return ctx.Tx.Exec(query, args...)
	}

	if ctx.Conn != nil {
		return ctx.Conn.ExecContext(c, query, args...)
	}

	conn, err := Connect(helperPostgresConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.ExecContext(c, query, args...)
}

func (ctx *DBContext) prepareContext(c context.Context, query string) (*sql.Stmt, error) {
	if err := ctx.validateResolution(); err != nil {
		return nil, err
	}

	if ctx.Tx != nil {
		if t, ok := ctx.Tx.(TxContext); ok {
			return t.PrepareContext(c, query)
		}
		return ctx.Tx.Prepare(query)
	}

	if ctx.Conn != nil {
		return ctx.Conn.PrepareContext(c, query)
	}

	conn, err := Connect(helperPostgresConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.PrepareContext(c, query)
}

func (ctx *DBContext) queryContext(c context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := ctx.validateResolution(); err != nil {
		return nil, err
	}

	if ctx.Tx != nil {
		if t, ok := ctx.Tx.(TxContext); ok {
			return t.QueryContext(c, query, args...)
		}
		return ctx.Tx.Query(query, args...)
	}

	if ctx.Conn != nil {
		return ctx.Conn.QueryContext(c, query, args...)
	}

	conn, err := Connect(helperPostgresConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.QueryContext(c, query, args...)
}

// Tx is the minimal transaction/connection surface used by DBContext.
// It matches *sql.Tx methods for easy compatibility and test mocking.
type Tx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

// TxContext is an optional extension for transactions that support context methods.
// *sql.Tx implements these.
type TxContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}
