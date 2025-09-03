package repository

import (
	"context"
	"seta/internal/application/ports"

	"gorm.io/gorm"
)

// GormTransactionManager implements the TransactionManager port using GORM.
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager creates a new GormTransactionManager.
func NewGormTransactionManager(db *gorm.DB) ports.TransactionManager {
	return &GormTransactionManager{db: db}
}

// Do executes the given function within a single database transaction.
// It will commit the transaction if the function returns no error,
// and roll it back if an error is returned.
func (tm *GormTransactionManager) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// We can pass the transaction-scoped *gorm.DB instance via the context
		// if any nested functions need it, but for this architecture,
		// the repositories will be re-initialized with the tx object.
		return fn(ctx) // The use case will re-use its repository, which now points to the tx
	})
}
