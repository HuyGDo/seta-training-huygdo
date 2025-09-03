package ports

import (
	"context"
	"io"
)

// UserImportSummary summarizes the result of a bulk user import.
type UserImportSummary struct {
	Succeeded int
	Failed    int
	Failures  []FailedRecord
}

// FailedRecord contains details about a single failed import record.
type FailedRecord struct {
	Record []string
	Reason string
}

// UserImporter defines the interface for importing users from a data source.
// This decouples the application's use case from the specific external service implementation.
type UserImporter interface {
	ImportUsers(ctx context.Context, file io.Reader) (UserImportSummary, error)
}
