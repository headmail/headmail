// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type txKey struct{}

// injectTx injects a transaction into the context.
func injectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx extracts a transaction from the context.
// If no transaction is found, it returns the original DB instance.
func extractTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

func getExistingTx(ctx context.Context, db *gorm.DB) (*gorm.DB, error) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		return nil, errors.New("no transaction")
	}
	return tx, nil
}
