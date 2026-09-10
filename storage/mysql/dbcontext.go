package mysql

import (
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

	// PrepareFn is an optional override for preparing SQL statements.
	PrepareFn func(query string) (*sql.Stmt, error)

	// QueryFn is an optional override for executing SQL queries that return rows.
	QueryFn func(query string, args ...any) (*sql.Rows, error)
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
	return ctx.exec(query, args...)
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
	return ctx.prepare(query)
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
	return ctx.query(query, args...)
}

// exec is the internal fallback for Exec(), used if ExecFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) exec(query string, args ...any) (sql.Result, error) {
	if ctx.Tx != nil {
		return ctx.Tx.Exec(query, args...)
	}
	if ctx.Conn != nil {
		return ctx.Conn.Exec(query, args...)
	}
	conn, err := Connect(helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.Exec(query, args...)
}

// prepare is the internal fallback for Prepare(), used if PrepareFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) prepare(query string) (*sql.Stmt, error) {
	if ctx.Tx != nil {
		return ctx.Tx.Prepare(query)
	}
	if ctx.Conn != nil {
		return ctx.Conn.Prepare(query)
	}
	conn, err := Connect(helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.Prepare(query)
}

// query is the internal fallback for Query(), used if QueryFn is nil.
// It chooses Tx → Conn → Cluster-based connection in that order.
func (ctx *DBContext) query(query string, args ...any) (*sql.Rows, error) {
	if ctx.Tx != nil {
		return ctx.Tx.Query(query, args...)
	}
	if ctx.Conn != nil {
		return ctx.Conn.Query(query, args...)
	}
	conn, err := Connect(helperMySqlConnector, ctx.Cluster)
	if err != nil {
		return nil, err
	}
	return conn.Query(query, args...)
}
