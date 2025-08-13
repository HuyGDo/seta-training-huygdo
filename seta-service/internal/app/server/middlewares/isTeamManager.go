package middlewares

import (
	"net/http"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IsTeamManager creates a gin middleware to check if a user is a manager of a team.
func IsTeamManager(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamIDStr := c.Param("teamId")
		teamID, err := uuid.Parse(teamIDStr)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
			return
		}

		userIDStr, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User ID not found in token"})
			return
		}
		
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
			return
		}

		var team models.Team
		if err := db.Preload("Managers").First(&team, "id = ?", teamID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}

		isManager := false
		for _, manager := range team.Managers {
			if manager.ID == userID {
				isManager = true
				break
			}
		}

		if !isManager {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not a manager of this team"})
			return
		}

		c.Next()
	}
}
