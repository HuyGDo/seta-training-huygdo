package controllers

import (
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userId")
	managerID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
		return
	}

	var manager models.User
	if err := tc.db.First(&manager, "id = ?", managerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	team := models.Team{
		TeamName: input.TeamName,
		Managers: []models.User{manager},
	}

	if err := tc.db.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
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
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var input AddRemoveMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memberID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var team models.Team
	if err := tc.db.Preload("Members").First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var member models.User
	if err := tc.db.First(&member, "id = ?", memberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Members").Append(&member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to team"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// RemoveMember removes a member from a team.
func (tc *TeamController) RemoveMember(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var member models.User
	if err := tc.db.First(&member, "id = ?", memberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Members").Delete(&member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member from team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// AddManager adds a manager to a team.
func (tc *TeamController) AddManager(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var input AddRemoveMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	managerID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var team models.Team
	if err := tc.db.Preload("Managers").First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var manager models.User
	if err := tc.db.First(&manager, "id = ?", managerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Managers").Append(&manager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add manager to team"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// RemoveManager removes a manager from a team.
func (tc *TeamController) RemoveManager(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	managerID, err := uuid.Parse(c.Param("managerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manager ID"})
		return
	}

	var team models.Team
	if err := tc.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var manager models.User
	if err := tc.db.First(&manager, "id = ?", managerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	if err := tc.db.Model(&team).Association("Managers").Delete(&manager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove manager from team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manager removed successfully"})
}

// GetTeamAssets retrieves all assets belonging to or shared with a team's members.
func (tc *TeamController) GetTeamAssets(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := tc.db.Preload("Members").First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	memberIDs := make([]uuid.UUID, len(team.Members))
	for i, member := range team.Members {
		memberIDs[i] = member.ID
	}

	var folders []models.Folder
	if err := tc.db.
		Joins("JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id IN (?) OR folder_shares.user_id IN (?)", memberIDs, memberIDs).
		Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve folders for the team"})
		return
	}

	var notes []models.Note
	if err := tc.db.
		Joins("JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("notes.owner_id IN (?) OR note_shares.user_id IN (?)", memberIDs, memberIDs).
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes for the team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"notes":   notes,
	})
}
