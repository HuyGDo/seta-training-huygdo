package controllers

import (
	"net/http"
	"seta/internal/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseController provides shared functionality for all controllers.
type BaseController struct {
    db *gorm.DB
}

// NewBaseController creates a new BaseController.
func NewBaseController(db *gorm.DB) BaseController {
    return BaseController{db: db}
}

// GetUserIDFromContext retrieves the user ID from the Gin context.
// It handles parsing and error responses automatically.
func (bc *BaseController) GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
    userIDStr, exists := c.Get("userId")
    if !exists {
        _ = c.Error(&errorHandling.CustomError{
            Code:    http.StatusUnauthorized,
            Message: "User not authenticated",
        })
        return uuid.Nil, false
    }

    userID, err := uuid.Parse(userIDStr.(string))
    if err != nil {
        _ = c.Error(&errorHandling.CustomError{
            Code:    http.StatusInternalServerError,
            Message: "Invalid user ID format",
        })
        return uuid.Nil, false
    }

    return userID, true
}