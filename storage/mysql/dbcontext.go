package mysql

import (
	"context"
	"database/sql"
)

// DBContext defines a contextual wrapper around SQL execution logic.
// It allows execution via an existing transaction (Tx), a manually provided
// database connection (Conn), or, if neither is set, establishes a connection
// using the provided cluster name.
//
// Custom execution behaviors can be injected via ExecFn, PrepareFn, and QueryFn
// to facilitate mocking or alternative behaviors (e.g., instrumentation).
type DBContext struct {
	// Tx represents an optional active SQL transaction. If set, all operations
	// will be executed using this transaction.
	Tx Tx

	// Conn is an optional MySQL connection interface. If set and Tx is nil,
	// this will be used to execute operations.
	Conn MySqlDbInterface

	// Cluster is a named MySQL cluster identifier. If both Tx and Conn are nil,
	// this name is used to fetch a new database connection via helperMySqlConnector.
	Cluster string

	// ExecFn is an optional override for executing SQL statements.
	// If set, it is invoked instead of the default internal logic.
	ExecFn func(query string, args ...any) (sql.Result, error)

	// ExecContextFn is an optional override for executing SQL statements with context.
	ExecContextFn func(ctx context.Context, query string, args ...any) (sql.Result, error)

	// PrepareFn is an optional override for preparing SQL statements.
	PrepareFn func(query string) (*sql.Stmt, error)

	// PrepareContextFn is an optional override for preparing SQL statements with context.
	PrepareContextFn func(ctx context.Context, query string) (*sql.Stmt, error)

	// QueryFn is an optional override for executing SQL queries that return rows.
	QueryFn func(query string, args ...any) (*sql.Rows, error)

	// QueryContextFn is an optional override for executing SQL queries with context.
	QueryContextFn func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Exec executes a SQL statement (e.g., INSERT, UPDATE) within the context.
//
// Priority:
//  1. ExecFn (if provided)
//  2. Tx.Exec() (if Tx is set)
//  3. Conn.Exec() (if Conn is set)
//  4. Connects via Cluster and uses Exec()
func (ctx *DBContext) Exec(query string, args ...any) (sql.Result, error) {
	if ctx.ExecFn != nil {
		return ctx.ExecFn(query, args...)
	}
	return ctx.execContext(context.Background(), query, args...)
}

// ExecContext executes a SQL statement (e.g., INSERT, UPDATE) with context.
func (ctx *DBContext) ExecContext(c context.Context, query string, args ...any) (sql.Result, error) {
	if ctx.ExecContextFn != nil {
		return ctx.ExecContextFn(c, query, args...)
	}
	return ctx.execContext(c, query, args...)
}

// Prepare prepares a SQL statement within the context.
//
// Priority:
//  1. PrepareFn (if provided)
//  2. Tx.Prepare() (if Tx is set)
//  3. Conn.Prepare() (if Conn is set)
//  4. Connects via Cluster and uses Prepare()
func (ctx *DBContext) Prepare(query string) (*sql.Stmt, error) {
	if ctx.PrepareFn != nil {
		return ctx.PrepareFn(query)
	}
	return ctx.prepareContext(context.Background(), query)
}

// PrepareContext prepares a SQL statement with context.
func (ctx *DBContext) PrepareContext(c context.Context, query string) (*sql.Stmt, error) {
	if ctx.PrepareContextFn != nil {
		return ctx.PrepareContextFn(c, query)
	}
	return ctx.prepareContext(c, query)
}

// Query executes a SQL query that returns rows.
//
// Priority:
//  1. QueryFn (if provided)
//  2. Tx.Query() (if Tx is set)
//  3. Conn.Query() (if Conn is set)
//  4. Connects via Cluster and uses Query()
func (ctx *DBContext) Query(query string, args ...any) (*sql.Rows, error) {
	if ctx.QueryFn != nil {
		return ctx.QueryFn(query, args...)
	}
	return ctx.queryContext(context.Background(), query, args...)
}

// QueryContext executes a SQL query that returns rows with context.
func (ctx *DBContext) QueryContext(c context.Context, query string, args ...any) (*sql.Rows, error) {
	if ctx.QueryContextFn != nil {
		return ctx.QueryContextFn(c, query, args...)
	}
	return ctx.queryContext(c, query, args...)
}

// exec is the internal fallback for Exec(), used if ExecFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) exec(query string, args ...any) (sql.Result, error) {
	return ctx.execContext(context.Background(), query, args...)
}

func (ctx *DBContext) execContext(c context.Context, query string, args ...any) (sql.Result, error) {
	if c == nil {
		c = context.Background()
	}
	if ctx.Tx != nil {
		if tx, ok := ctx.Tx.(TxContext); ok {
			return tx.ExecContext(c, query, args...)
		}
		return ctx.Tx.Exec(query, args...)
	}
	if ctx.Conn != nil {
		return ctx.Conn.ExecContext(c, query, args...)
	}
	conn, err := ConnectContext(c, helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.ExecContext(c, query, args...)
}

// prepare is the internal fallback for Prepare(), used if PrepareFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) prepare(query string) (*sql.Stmt, error) {
	return ctx.prepareContext(context.Background(), query)
}

func (ctx *DBContext) prepareContext(c context.Context, query string) (*sql.Stmt, error) {
	if c == nil {
		c = context.Background()
	}
	if ctx.Tx != nil {
		if tx, ok := ctx.Tx.(TxContext); ok {
			return tx.PrepareContext(c, query)
		}
		return ctx.Tx.Prepare(query)
	}
	if ctx.Conn != nil {
		return ctx.Conn.PrepareContext(c, query)
	}
	conn, err := ConnectContext(c, helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.PrepareContext(c, query)
}

// query is the internal fallback for Query(), used if QueryFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) query(query string, args ...any) (*sql.Rows, error) {
	return ctx.queryContext(context.Background(), query, args...)
}

func (ctx *DBContext) queryContext(c context.Context, query string, args ...any) (*sql.Rows, error) {
	if c == nil {
		c = context.Background()
	}
	if ctx.Tx != nil {
		if tx, ok := ctx.Tx.(TxContext); ok {
			return tx.QueryContext(c, query, args...)
		}
		return ctx.Tx.Query(query, args...)
	}
	if ctx.Conn != nil {
		return ctx.Conn.QueryContext(c, query, args...)
	}
	conn, err := ConnectContext(c, helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.QueryContext(c, query, args...)
}
