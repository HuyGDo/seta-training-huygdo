package models

import (
	"time"

	"github.com/google/uuid"
)

// gormTeam is the GORM model for a team. It includes gorm tags for database mapping.
type GormTeam struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	TeamName  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Managers  []GormTeamManager `gorm:"foreignKey:TeamID"`
	Members   []GormTeamMember  `gorm:"foreignKey:TeamID"`
}

func (GormTeam) TableName() string { return "teams" }

// gormTeamManager is the GORM model for the team_managers join table.
type GormTeamManager struct {
	TeamID uuid.UUID `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"primaryKey"`
	IsLead bool      `gorm:"default:false"`
}

func (GormTeamManager) TableName() string { return "team_managers" }

// gormTeamMember is the GORM model for the team_members join table.
type GormTeamMember struct {
	TeamID uuid.UUID `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"primaryKey"`
}

func (GormTeamMember) TableName() string { return "team_members" }
