// internal/application/ports/auth.go
package ports

import (
	"context"
	"seta/internal/domain/common"
)

// User represents a user retrieved from an external authentication service.
type User struct {
	ID   common.UserID
	Role string // e.g., "MANAGER", "MEMBER"
}

// AuthService defines the interface for interacting with an external user/auth service.
// Its responsibility is AUTHENTICATION (verifying a user's identity).
type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*User, error)
	GetUser(ctx context.Context, userID common.UserID) (*User, error)
}

