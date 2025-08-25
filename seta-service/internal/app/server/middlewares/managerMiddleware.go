package middlewares

import (
	"errors"
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"
	"seta/internal/pkg/utils"

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

		var teamManager models.TeamManager
		err = db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&teamManager).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not a manager of this team"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify team manager status"})
			return
		}

		c.Next()
	}
}

func IsLeadManager(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        teamID, err := utils.GetUUIDFromParam(c, "teamId")
		if err != nil {
			_ = c.Error(err) 
			c.Abort()
			return
		}

		userID, err := utils.GetUserUUIDFromContext(c)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}


        var manager models.TeamManager
        err = db.Where("team_id = ? AND user_id = ? AND is_lead = ?", teamID, userID, true).First(&manager).Error
        if err != nil {
            _ = c.Error(&errorHandling.CustomError{
                Code: http.StatusForbidden, 
                Message: "You must be a lead manager to perform this action",
            })
            c.Abort() 
            return
        }
        c.Next()
    }
}