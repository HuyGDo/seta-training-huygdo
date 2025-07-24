package models

import (
	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username     string
	Email        string `gorm:"uniqueIndex"`
	Role         string `gorm:"type:enum('manager','member')"`
	PasswordHash string
}
