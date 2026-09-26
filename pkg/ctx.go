package pkg

import (
	"context"
	"database/sql"
)

type CtxKey string

var TxCtxKey = CtxKey("tx")

func txFromCtx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(TxCtxKey).(*sql.Tx)
	return tx, ok
}

func ctxWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, TxCtxKey, tx)
}
