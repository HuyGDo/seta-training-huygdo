package controllers

import (
	"context"
	"net/http"
	"seta/internal/pkg/database"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/kafka"
	"seta/internal/pkg/models"
	"seta/internal/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamController now has its own db field and no longer embeds BaseController.
type TeamController struct {
	db *gorm.DB
}

// NewTeamController creates a new TeamController, injecting the db dependency.
func NewTeamController(db *gorm.DB) *TeamController {
	return &TeamController{db: db}
}

type ManagerInput struct {
	ManagerID   uuid.UUID `json:"managerId" binding:"required"`
	ManagerName string    `json:"managerName"`
	IsLead      bool      `json:"isLead"`
}

type MemberInput struct {
	MemberID   uuid.UUID `json:"memberId" binding:"required"`
	MemberName string    `json:"memberName"`
}

type CreateTeamInput struct {
	TeamName string         `json:"teamName" binding:"required"`
	Managers []ManagerInput `json:"managers" binding:"required,min=1"`
	Members  []MemberInput  `json:"members"`
}

// CreateTeam creates a new team.
func (tc *TeamController) CreateTeam(c *gin.Context) {
	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body: " + err.Error()})
		return
	}

	creatorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var leadManagerCount int
	var isCreatorAManager bool
	for _, manager := range input.Managers {
		if manager.ManagerID == creatorUserID {
			isCreatorAManager = true
		}
		if manager.IsLead {
			leadManagerCount++
		}
	}

	// Validation: Ensure the creator is in the manager list
	if !isCreatorAManager {
		_ = c.Error(&errorHandling.CustomError{
			Code:    http.StatusBadRequest,
			Message: "The user creating the team must be included in the managers list.",
		})
		return
	}

	// Validation: Ensure there is exactly one lead manager
	if leadManagerCount != 1 {
		_ = c.Error(&errorHandling.CustomError{
			Code:    http.StatusBadRequest,
			Message: "Exactly one manager must be designated as the lead (isLead: true).",
		})
		return
	}

	team := models.Team{TeamName: input.TeamName}

	err = tc.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&team).Error; err != nil {
			return err
		}
		for _, manager := range input.Managers {
			teamManager := models.TeamManager{TeamID: team.ID, UserID: manager.ManagerID, IsLead: manager.IsLead}
			if err := tx.Create(&teamManager).Error; err != nil {
				return err
			}
		}
		for _, member := range input.Members {
			teamMember := models.TeamMember{TeamID: team.ID, UserID: member.MemberID}
			if err := tx.Create(&teamMember).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create team: " + err.Error()})
		return
	}
	
	go kafka.ProduceTeamEvent(context.Background(), kafka.EventPayload{
		EventType: "TEAM_CREATED",
		TeamID:    team.ID.String(),
		ActionBy:  creatorUserID.String(),
	})


	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"team":    team,
	})
}

type AddRemoveMemberInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

// AddMember adds a member to a team.
func (tc *TeamController) AddMember(c *gin.Context) {
	teamID, err := utils.GetUUIDFromParam(c, "teamId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	var input AddRemoveMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	teamMember := models.TeamMember{TeamID: teamID, UserID: input.UserID}
	if err := tc.db.WithContext(c.Request.Context()).Create(&teamMember).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add member to team"})
		return
	}

	actorUserID, _ := utils.GetUserUUIDFromContext(c) // Error already handled by auth middleware
	go kafka.ProduceTeamEvent(context.Background(), kafka.EventPayload{
		EventType:    "MEMBER_ADDED",
		TeamID:       teamID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: input.UserID.String(),
	})

	c.Status(http.StatusNoContent)
}

// RemoveMember removes a member from a team.
func (tc *TeamController) RemoveMember(c *gin.Context) {
	teamID, err := utils.GetUUIDFromParam(c, "teamId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	memberID, err := utils.GetUUIDFromParam(c, "memberId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := tc.db.WithContext(c.Request.Context()).Delete(&models.TeamMember{TeamID: teamID, UserID: memberID}).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove member from team"})
		return
	}

	actorUserID, _ := utils.GetUserUUIDFromContext(c)
	go kafka.ProduceTeamEvent(context.Background(), kafka.EventPayload{
		EventType:    "MEMBER_REMOVED",
		TeamID:       teamID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: memberID.String(),
	})

	c.Status(http.StatusNoContent)
}

// AddManager adds a manager to a team.
func (tc *TeamController) AddManager(c *gin.Context) {
	teamID, err := utils.GetUUIDFromParam(c, "teamId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	var input AddRemoveMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	teamManager := models.TeamManager{TeamID: teamID, UserID: input.UserID}
	if err := tc.db.WithContext(c.Request.Context()).Create(&teamManager).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add manager to team"})
		return
	}

	actorUserID, _ := utils.GetUserUUIDFromContext(c)
	go kafka.ProduceTeamEvent(context.Background(), kafka.EventPayload{
		EventType:    "MANAGER_ADDED",
		TeamID:       teamID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: input.UserID.String(),
	})

	c.Status(http.StatusNoContent)
}

// RemoveManager removes a manager from a team.
func (tc *TeamController) RemoveManager(c *gin.Context) {
	teamID, err := utils.GetUUIDFromParam(c, "teamId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	managerID, err := utils.GetUUIDFromParam(c, "managerId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := tc.db.WithContext(c.Request.Context()).Delete(&models.TeamManager{TeamID: teamID, UserID: managerID}).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove manager from team"})
		return
	}

	actorUserID, _ := utils.GetUserUUIDFromContext(c)
	go kafka.ProduceTeamEvent(context.Background(), kafka.EventPayload{
		EventType:    "MANAGER_REMOVED",
		TeamID:       teamID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: managerID.String(),
	})

	c.Status(http.StatusNoContent)
}

// GetTeamAssets retrieves all assets belonging to or shared with a team's members.
func (tc *TeamController) GetTeamAssets(c *gin.Context) {
	teamID, err := utils.GetUUIDFromParam(c, "teamId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	cacheKey := "team:" + teamID.String() + ":members"
	var memberIDs []uuid.UUID

	// 1. Check Cache First (Cache-Aside)
	cachedMemberIDs, err := database.Rdb.SMembers(ctx, cacheKey).Result()
	if err == nil && len(cachedMemberIDs) > 0 {
		// Cache Hit
		for _, idStr := range cachedMemberIDs {
			id, _ := uuid.Parse(idStr)
			memberIDs = append(memberIDs, id)
		}
	} else {
		// Cache Miss
		if err := tc.db.Model(&models.TeamMember{}).Where("team_id = ?", teamID).Pluck("user_id", &memberIDs).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve team members"})
			return
		}

		// 4. Populate Cache
		if len(memberIDs) > 0 {
			// Convert UUIDs to strings for Redis
			stringIDs := make([]interface{}, len(memberIDs))
			for i, id := range memberIDs {
				stringIDs[i] = id.String()
			}
			database.Rdb.SAdd(ctx, cacheKey, stringIDs...)
			database.Rdb.Expire(ctx, cacheKey, 24*time.Hour) // Set expiration
		}
	}

	if len(memberIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"folders": []models.Folder{}, "notes": []models.Note{}})
		return
	}

	var assets struct {
		Folders []models.Folder `json:"folders"`
		Notes   []models.Note   `json:"notes"`
	}

	if err := tc.db.Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id IN (?) OR folder_shares.user_id IN (?)", memberIDs, memberIDs).
		Group("folders.folder_id").
		Find(&assets.Folders).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve folders"})
		return
	}

	if err := tc.db.Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("notes.owner_id IN (?) OR note_shares.user_id IN (?)", memberIDs, memberIDs).
		Group("notes.note_id").
		Find(&assets.Notes).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve notes"})
		return
	}

	c.JSON(http.StatusOK, assets)
}