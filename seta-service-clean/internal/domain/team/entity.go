package team

import (
	"seta/internal/domain/common"
	"time"
)

// Team is the aggregate root for the team domain.
// It represents a collection of users working together.
type Team struct {
	ID        common.TeamID
	Name      string
	Managers  []Manager
	Members   []Member
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Manager represents a user with management permissions within a team.
type Manager struct {
	UserID common.UserID
	IsLead bool
}

// Member represents a user who is part of a team.
type Member struct {
	UserID common.UserID
}
