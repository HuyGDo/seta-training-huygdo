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
	db *gorm.DB
}

// NewTeamController creates a new TeamController.
func NewTeamController(db *gorm.DB) *TeamController {
	return &TeamController{db: db}
}

// CreateTeamInput represents the input for creating a team.
type CreateTeamInput struct {
	TeamName string `json:"teamName" binding:"required"`
}

// CreateTeam creates a new team.
func (tc *TeamController) CreateTeam(c *gin.Context) {
	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	team := models.Team{TeamName: input.TeamName}
	if err := tc.db.WithContext(c.Request.Context()).Create(&team).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create team"})
		return
	}

	// FIX: Use the userID directly to create the association without a DB lookup.
	if err := tc.db.WithContext(c.Request.Context()).Model(&team).Association("Managers").Append(&models.User{ID: userID}); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add manager to team"})
		return
	}

	c.JSON(http.StatusCreated, team)
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

	var team models.Team
	if err := tc.db.WithContext(c.Request.Context()).First(&team, "id = ?", teamID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	// FIX: Removed the unnecessary user lookup from this database.
	userToAdd := &models.User{ID: input.UserID}
	if err := tc.db.WithContext(c.Request.Context()).Model(&team).Association("Members").Append(userToAdd); err != nil {
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

	var team models.Team
	if err := tc.db.WithContext(c.Request.Context()).First(&team, "id = ?", teamID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	// FIX: Removed the unnecessary user lookup from this database.
	memberToRemove := &models.User{ID: memberID}
	if err := tc.db.WithContext(c.Request.Context()).Model(&team).Association("Members").Delete(memberToRemove); err != nil {
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

	var team models.Team
	if err := tc.db.WithContext(c.Request.Context()).First(&team, "id = ?", teamID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	// FIX: Removed the unnecessary user lookup from this database.
	managerToAdd := &models.User{ID: input.UserID}
	if err := tc.db.WithContext(c.Request.Context()).Model(&team).Association("Managers").Append(managerToAdd); err != nil {
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

	var team models.Team
	if err := tc.db.WithContext(c.Request.Context()).First(&team, "id = ?", teamID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	// FIX: Removed the unnecessary user lookup from this database.
	managerToRemove := &models.User{ID: managerID}
	if err := tc.db.WithContext(c.Request.Context()).Model(&team).Association("Managers").Delete(managerToRemove); err != nil {
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
	if err := tc.db.Table("team_members").Where("team_id = ?", teamID).Pluck("user_id", &memberIDs).Error; err != nil {
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