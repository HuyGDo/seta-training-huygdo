package controllers

import (
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"
	"strconv"

	"github.com/gin-gonic/gin"
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
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	var user models.User
	if err := tc.db.First(&user, userID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "User not found"})
		return
	}

	team := models.Team{TeamName: input.Name}
	if err := tc.db.Create(&team).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create team"})
		return
	}

	if err := tc.db.Model(&team).Association("Managers").Append(&user); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add manager to team"})
		return
	}

	c.JSON(http.StatusCreated, team)
}

// AddRemoveMemberInput represents the input for adding or removing a team member.
type AddRemoveMemberInput struct {
	UserID string `json:"userId" binding:"required"`
}

// AddMember adds a member to a team.
func (tc *TeamController) AddMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	var input struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, uint(teamID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	var user models.User
	if err := tc.db.First(&user, input.UserID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "User not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Members").Append(&user); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add member to team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveMember removes a member from a team.
func (tc *TeamController) RemoveMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid member ID"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, uint(teamID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	var member models.User
	if err := tc.db.First(&member, uint(memberID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Member not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Members").Delete(&member); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove member from team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddManager adds a manager to a team.
func (tc *TeamController) AddManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	var input struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, uint(teamID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	var user models.User
	if err := tc.db.First(&user, input.UserID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "User not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Managers").Append(&user); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to add manager to team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveManager removes a manager from a team.
func (tc *TeamController) RemoveManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	managerIDStr := c.Param("managerId")
	managerID, err := strconv.ParseUint(managerIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid manager ID"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, uint(teamID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	var manager models.User
	if err := tc.db.First(&manager, uint(managerID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Manager not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Managers").Delete(&manager); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to remove manager from team"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTeamAssets retrieves all assets belonging to or shared with a team's members.
func (tc *TeamController) GetTeamAssets(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID"})
		return
	}

	var team models.Team
	if err := tc.db.Preload("Members").First(&team, uint(teamID)).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Team not found"})
		return
	}

	var assets struct {
		Folders []models.Folder `json:"folders"`
		Notes   []models.Note   `json:"notes"`
	}

	for _, member := range team.Members {
		var memberFolders []models.Folder
		if err := tc.db.Where("user_id = ?", member.ID).Find(&memberFolders).Error; err == nil {
			assets.Folders = append(assets.Folders, memberFolders...)
		}

		var memberNotes []models.Note
		if err := tc.db.Where("user_id = ?", member.ID).Find(&memberNotes).Error; err == nil {
			assets.Notes = append(assets.Notes, memberNotes...)
		}
	}

	c.JSON(http.StatusOK, assets)
}
