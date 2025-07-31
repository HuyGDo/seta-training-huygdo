package models

import "github.com/google/uuid"

// Team represents a team in the system.
type Team struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamName string
	Managers []User `gorm:"many2many:team_managers;"`
	Members  []User `gorm:"many2many:team_members;"`
}
