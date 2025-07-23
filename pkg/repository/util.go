package repository

import (
	"context"
)

func Transactional0(db DB, ctx context.Context, fn func(txCtx context.Context) error) error {
	txCtx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	err = fn(txCtx)
	if err != nil {
		_ = db.Rollback(txCtx)
	} else {
		err = db.Commit(txCtx)
	}
	return err
}

func Transactional1[T interface{}](db DB, ctx context.Context, fn func(txCtx context.Context) (T, error)) (T, error) {
	var out T
	txCtx, err := db.Begin(ctx)
	if err != nil {
		return out, err
	}
	out, err = fn(txCtx)
	if err != nil {
		_ = db.Rollback(txCtx)
	} else {
		err = db.Commit(txCtx)
	}
	return out, err
}
