package utils

import (
	"net/http"
	"seta/internal/domain/common"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserIDFromContext extracts and parses the user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (common.UserID, bool) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User ID not found in context"})
		return common.UserID(uuid.Nil), false
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format in context"})
		return common.UserID(uuid.Nil), false
	}
	return common.UserID(userID), true
}

// GetUUIDFromParam extracts and parses a UUID from a URL parameter.
func GetUUIDFromParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	idStr := c.Param(paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid UUID format for parameter: " + paramName})
		return uuid.Nil, false
	}
	return id, true
}
