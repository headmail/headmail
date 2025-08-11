package repository

import (
	"context"
)

type txContextKey struct{}

type txContextData struct {
	needRollback bool
}

func transactionWrap(db DB, ctx context.Context, fn func(txCtx context.Context) error) error {
	var err error
	var newTx bool
	txData, ok := ctx.Value(txContextKey{}).(*txContextData)
	if !ok {
		newTx = true
		txData = &txContextData{}
		ctx = context.WithValue(ctx, txContextKey{}, txData)
		ctx, err = db.Begin(ctx)
		if err != nil {
			return err
		}
	}

	err = fn(ctx)
	if err != nil {
		txData.needRollback = true
	}
	if newTx {
		if err != nil || txData.needRollback {
			_ = db.Rollback(ctx)
			return err
		}
		return db.Commit(ctx)
	}
	return err
}

func Transactional0(db DB, ctx context.Context, fn func(txCtx context.Context) error) error {
	return transactionWrap(db, ctx, fn)
}

func Transactional1[T interface{}](db DB, ctx context.Context, fn func(txCtx context.Context) (T, error)) (T, error) {
	var out T
	err := transactionWrap(db, ctx, func(txCtx context.Context) error {
		var e error
		out, e = fn(txCtx)
		return e
	})
	return out, err
}
