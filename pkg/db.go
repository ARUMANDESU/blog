package pkg

import "context"

type TxManager interface {
	InTx(context.Context, func(context.Context) error) error
}

type NoOpTxManager struct{}

func (m *NoOpTxManager) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type SqliteTxManager struct {
}

func (m *SqliteTxManager) InTx(ctx context.Context, fn func(context.Context) error) error {
	// TODO: implement this
	return fn(ctx)
}
