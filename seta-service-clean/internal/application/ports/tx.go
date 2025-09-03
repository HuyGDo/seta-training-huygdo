package ports

import "context"

// TransactionManager defines an interface for managing database transactions.
// This allows use cases to wrap operations in a transaction without depending on a specific DB implementation.
type TransactionManager interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}