package pkg

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNestedTxs = errors.New("nested txs")

type TxManager interface {
	// InTx begins tx runs and propagates through context into provided func.
	// Nesting is forbiden, else [ErrNestedTxs] is returned.
	InTx(context.Context, func(context.Context) error) error
}

type NoOpTxManager struct{}

func (m *NoOpTxManager) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type SqlTxManager struct {
	db        *sql.DB
	txOptions *sql.TxOptions
}

func NewSqlTxManager(db *sql.DB, opts *sql.TxOptions) *SqlTxManager {
	return &SqlTxManager{db: db, txOptions: opts}
}

func (m *SqlTxManager) InTx(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := txFromCtx(ctx); ok {
		return ErrNestedTxs
	}

	tx, err := m.db.BeginTx(ctx, m.txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = fn(ctxWithTx(ctx, tx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func SqlConn(ctx context.Context, db *sql.DB) DBTX {
	if tx, ok := txFromCtx(ctx); ok {
		return tx
	}
	return db
}
