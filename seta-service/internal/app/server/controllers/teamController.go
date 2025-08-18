package controllers

import (
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamController handles team-related requests.
type TeamController struct {
	BaseController // Embed the BaseController
}

// NewTeamController creates a new TeamController.
func NewTeamController(db *gorm.DB) *TeamController {
	// Initialize the embedded BaseController
	return &TeamController{BaseController: NewBaseController(db)}
}

// CreateTeamInput represents the input for creating a team.

type ManagerInput struct {
	ManagerID   uuid.UUID `json:"managerId" binding:"required"`
	ManagerName string    `json:"managerName"` // Name is optional, we only need the ID
}

type MemberInput struct {
	MemberID   uuid.UUID `json:"memberId" binding:"required"`
	MemberName string    `json:"memberName"` // Name is optional
}

// CreateTeamInput now matches your desired request body
type CreateTeamInput struct {
	TeamName string         `json:"teamName" binding:"required"`
	Managers []ManagerInput `json:"managers" binding:"required,min=1"` // Require at least one manager
	Members  []MemberInput  `json:"members"`
}

// CreateTeam creates a new team and adds the creator as the first manager.
func (tc *TeamController) CreateTeam(c *gin.Context) {
	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body: " + err.Error()})
		return
	}

	// This helper is still useful for getting the creator's ID,
	// but the creator might not be in the managers list, so we handle that.
	creatorUserID, ok := tc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response
	}

	// Check if the creator is in the provided managers list.
	isCreatorAManager := false
	for _, manager := range input.Managers {
		if manager.ManagerID == creatorUserID {
			isCreatorAManager = true
			break
		}
	}

	if !isCreatorAManager {
		_ = c.Error(&errorHandling.CustomError{
			Code:    http.StatusBadRequest,
			Message: "The user creating the team must be included in the managers list.",
		})
		return
	}

	team := models.Team{TeamName: input.TeamName}

	// Use a transaction to create the team, managers, and members atomically
	err := tc.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		// 1. Create the team
		if err := tx.Create(&team).Error; err != nil {
			return err
		}

		// 2. Create the managers from the input list
		for _, manager := range input.Managers {
			teamManager := models.TeamManager{TeamID: team.ID, UserID: manager.ManagerID}
			if err := tx.Create(&teamManager).Error; err != nil {
				return err
			}
		}

		// 3. Create the members from the input list
		for _, member := range input.Members {
			teamMember := models.TeamMember{TeamID: team.ID, UserID: member.MemberID}
			if err := tx.Create(&teamMember).Error; err != nil {
				return err
			}
		}

		return nil // Commit transaction
	})

	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create team: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"team":    team,
	})
}

// AddRemoveMemberInput represents the input for adding or removing a team member.
type AddRemoveMemberInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

// AddMember adds a member to a team.
func (tc *TeamController) AddMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
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

	c.Status(http.StatusNoContent)
}

// RemoveMember removes a member from a team.
func (tc *TeamController) RemoveMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid member ID"})
		return
	}

	if err := tc.db.WithContext(c.Request.Context()).Delete(&models.TeamMember{TeamID: teamID, UserID: memberID}).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove member from team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddManager adds a manager to a team.
func (tc *TeamController) AddManager(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
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

	c.Status(http.StatusNoContent)
}

// RemoveManager removes a manager from a team.
func (tc *TeamController) RemoveManager(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	managerID, err := uuid.Parse(c.Param("managerId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid manager ID"})
		return
	}

	if err := tc.db.WithContext(c.Request.Context()).Delete(&models.TeamManager{TeamID: teamID, UserID: managerID}).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove manager from team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTeamAssets retrieves all assets belonging to or shared with a team's members.
func (tc *TeamController) GetTeamAssets(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	var memberIDs []uuid.UUID
	if err := tc.db.Model(&models.TeamMember{}).Where("team_id = ?", teamID).Pluck("user_id", &memberIDs).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve team members"})
		return
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