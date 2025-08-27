package utils

import (
	"fmt"
	"net/http"
	"seta/internal/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserUUIDFromContext retrieves the user ID from the Gin context and parses it.
// It returns a proper error that can be handled by the caller.
func GetUserUUIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userId")
	if !exists {
		return uuid.Nil, &errorHandling.CustomError{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
		}
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, &errorHandling.CustomError{
			Code:    http.StatusInternalServerError,
			Message: "Invalid user ID format in token",
		}
	}

	return userID, nil
}

// GetUUIDFromParam retrieves an ID from a URL parameter and parses it.
func GetUUIDFromParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return uuid.Nil, &errorHandling.CustomError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Missing URL parameter: %s", paramName),
		}
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, &errorHandling.CustomError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid UUID format for parameter: %s", paramName),
		}
	}

	return id, nil
}