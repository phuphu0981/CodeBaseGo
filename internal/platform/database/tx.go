package database

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type txContextKey struct{}

// WithTransaction executes fn inside a GORM transaction.
// If the context already has an ongoing transaction, fn is called with the current context (joins the transaction).
// If fn returns an error or panics, the transaction is rolled back; otherwise it is committed.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(txCtx context.Context) error, opts ...*sql.TxOptions) error {
	if db == nil {
		return fmt.Errorf("db cannot be nil")
	}

	// If already in a transaction, reuse the context without opening a nested transaction
	if _, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txContextKey{}, tx)
		return fn(txCtx)
	}, opts...)
}

// GetDB returns the transaction *gorm.DB from ctx if one exists, otherwise returns defaultDB.WithContext(ctx).
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	if defaultDB != nil {
		return defaultDB.WithContext(ctx)
	}
	return nil
}
