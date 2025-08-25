package models

import "github.com/google/uuid"

// Team represents a team in the system.
type Team struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:id"`
	TeamName string

}

type TeamManager struct {
    TeamID uuid.UUID `gorm:"primaryKey"`
    UserID uuid.UUID `gorm:"primaryKey"`
    IsLead bool `gorm:"default:false"`
}

func (TeamManager) TableName() string {
    return "team_managers"
}

// TeamMember represents the join table between teams and members.
type TeamMember struct {
    TeamID uuid.UUID `gorm:"primaryKey"`
    UserID uuid.UUID `gorm:"primaryKey"`
}

func (TeamMember) TableName() string {
    return "team_members"
}