package sqlmap

import (
	"context"
	"database/sql"
)

// DB wraps a *sql.DB object.
type DB interface {
	Queryer
	Execer
	Preparer
	Pinger
	Beginner
	Closer

	DB() *sql.DB
}

// Tx wraps a *sql.Tx object.
type Tx interface {
	Queryer
	Execer
	Transactor

	StmtContext(context.Context, Stmt) Stmt

	Tx() *sql.Tx
}

// Stmt wraps a *sql.Stmt object.
type Stmt interface {
	PreparedQueryer
	PreparedExecer

	Stmt() *sql.Stmt
}

// ClassicQueryer represents the object that can perform SQL queries.
type ClassicQueryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// ContextQueryer represents the object that can perform SQL queries.
type ContextQueryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Queryer represents the object that can perform SQL queries.
type Queryer interface {
	ClassicQueryer
	ContextQueryer
}

// ClassicPreparedQueryer represents the object that can perform prepared SQL queries.
type ClassicPreparedQueryer interface {
	Query(args ...interface{}) (*sql.Rows, error)
	QueryRow(args ...interface{}) *sql.Row
}

// ContextPreparedQueryer represents the object that can perform prepared SQL queries.
type ContextPreparedQueryer interface {
	QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
}

// PreparedQueryer represents the object that can perform prepared SQL queries.
type PreparedQueryer interface {
	ClassicPreparedQueryer
	ContextPreparedQueryer
}

// ClassicExecer represents the object that can execute SQL queries.
type ClassicExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ContextExecer represents the object that can execute SQL queries.
type ContextExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Execer represents the object that can execute SQL queries.
type Execer interface {
	ClassicExecer
	ContextExecer
}

// ClassicPreparedExecer represents the object that can execute prepared SQL queries.
type ClassicPreparedExecer interface {
	Exec(args ...interface{}) (sql.Result, error)
}

// ContextPreparedExecer represents the object that can execute prepared SQL queries.
type ContextPreparedExecer interface {
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
}

// PreparedExecer represents the object that can execute prepared SQL queries.
type PreparedExecer interface {
	ClassicPreparedExecer
	ContextPreparedExecer
}

// ClassicPreparer represents the object that can create SQL prepared statements.
type ClassicPreparer interface {
	Prepare(query string) (Stmt, error)
}

// ContextPreparer represents the object that can create SQL prepared statements.
type ContextPreparer interface {
	PrepareContext(ctx context.Context, query string) (Stmt, error)
}

// Preparer represents the object that can create SQL prepared statements.
type Preparer interface {
	ClassicPreparer
	ContextPreparer
}

// ClassicPinger represents the object that can verify a connection.
type ClassicPinger interface {
	Ping() error
}

// ContextPinger represents the object that can verify a connection.
type ContextPinger interface {
	PingContext(ctx context.Context) error
}

// Pinger represents the object that can verify a connection.
type Pinger interface {
	ClassicPinger
	ContextPinger
}

// ClassicBeginner represents the object that can begin a transaction.
type ClassicBeginner interface {
	Begin() (Tx, error)
}

// ContextBeginner represents the object that can begin a transaction.
type ContextBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (Tx, error)
}

// Beginner represents the object that can begin a transaction.
type Beginner interface {
	ClassicBeginner
	ContextBeginner
}

// Transactor represents the object that can commit or rollback a transaciton.
type Transactor interface {
	Commit() error
	Rollback() error
}

// Closer represents the object that can close a connection.
type Closer interface {
	Close() error
}
