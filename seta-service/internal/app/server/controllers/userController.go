package controllers

import (
	"net/http"
	"seta/internal/app/server/services"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserController handles user-related HTTP requests.
type UserController struct {
	BaseController // Embed BaseController for DB access and helpers
	userService    *services.UserService
}

// NewUserController creates a new UserController.
func NewUserController(db *gorm.DB, userService *services.UserService) *UserController {
	return &UserController{
		BaseController: NewBaseController(db),
		userService:    userService,
	}
}

// ImportUsers handles the file upload and calls the user service to process it.
func (uc *UserController) ImportUsers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "File not provided in 'file' form field"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to open uploaded file"})
		return
	}
	defer openedFile.Close()

	// C.1: Context Propagation - Pass the request context to the service.
	summary, err := uc.userService.ImportUsers(c.Request.Context(), openedFile)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	// B.2: Detailed Failure Reasons - Return the detailed summary.
	c.JSON(http.StatusOK, gin.H{
		"message":   "User import process completed.",
		"succeeded": summary.Succeeded,
		"failed":    summary.Failed,
		"failures":  summary.Failures,
	})
}




// GetUserAssets retrieves all assets owned by or shared with a specific user.
func (uc *UserController) GetUserAssets(c *gin.Context) {
	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid user ID"})
		return
	}

	// Use the helper from BaseController to get the authenticated user's ID
	authUserID, ok := uc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response
	}

	// Authorization check
	if authUserID != targetUserID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to view these assets"})
		return
	}

	var folders []models.Folder
	if err := uc.db.WithContext(c.Request.Context()).
		Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id = ? OR folder_shares.user_id = ?", targetUserID, targetUserID).
		Group("folders.folder_id").
		Find(&folders).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve folders for the user"})
		return
	}

	var notes []models.Note
	if err := uc.db.WithContext(c.Request.Context()).
		Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Joins("LEFT JOIN folder_shares ON notes.folder_id = folder_shares.folder_id").
		Where("notes.owner_id = ? OR note_shares.user_id = ? OR folder_shares.user_id = ?", targetUserID, targetUserID, targetUserID).
		Group("notes.note_id").
		Find(&notes).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve notes for the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"notes":   notes,
	})
}