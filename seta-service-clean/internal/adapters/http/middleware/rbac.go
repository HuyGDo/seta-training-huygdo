package middleware

import (
	"errors"
	"net/http"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IsAuthorizedRole checks if the user's role from the JWT matches one of the allowed roles.
// IsAuthorizedRole checks if the user's role from the token is one of the allowed roles.
func IsAuthorizedRole(authorizedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "User role not found in token"})
			c.Abort()
			return
		}

		userRole := role.(string)
		isAuthorized := false
		for _, authorizedRole := range authorizedRoles {
			if userRole == authorizedRole {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsTeamManagerMiddleware uses the TeamRepository port for authorization.
func IsTeamManagerMiddleware(teamRepo ports.TeamRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamIDStr := c.Param("teamId")
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID format"})
			c.Abort()
			return
		}

		userIDStr, _ := c.Get("userId")
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID in context"})
			c.Abort()
			return
		}

		isManager, err := teamRepo.IsManager(c.Request.Context(), common.TeamID(teamID), common.UserID(userID))
		if err != nil {
			// Check for a specific "not found" error if your repo returns one, otherwise handle generically
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not a manager of this team"})
			} else {
				c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to verify team manager status"})
			}
			c.Abort()
			return
		}

		if !isManager {
			c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not a manager of this team"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsLeadManagerMiddleware uses the TeamRepository port for lead-specific authorization.
func IsLeadManagerMiddleware(teamRepo ports.TeamRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamIDStr := c.Param("teamId")
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid team ID format"})
			c.Abort()
			return
		}

		userIDStr, _ := c.Get("userId")
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID in context"})
			c.Abort()
			return
		}

		isLead, err := teamRepo.IsLeadManager(c.Request.Context(), common.TeamID(teamID), common.UserID(userID))
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to verify lead manager status"})
			c.Abort()
			return
		}

		if !isLead {
			c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You must be a lead manager to perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}