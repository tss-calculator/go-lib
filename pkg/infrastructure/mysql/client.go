package mysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type ClientContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Transaction interface {
	ClientContext
	Commit() error
	Rollback() error
}

type TransactionalClient interface {
	ClientContext
	BeginTransaction() (Transaction, error)
}

type transactionalClient struct {
	*sqlx.DB
}

func (client *transactionalClient) BeginTransaction() (Transaction, error) {
	return client.Beginx()
}
