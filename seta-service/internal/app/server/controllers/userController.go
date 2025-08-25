package controllers

import (
	"net/http"
	"seta/internal/app/server/services"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"
	"seta/internal/pkg/utils" // Import the new utils package

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController handles user-related HTTP requests.
type UserController struct {
	db          *gorm.DB
	userService *services.UserService
}

// NewUserController creates a new UserController.
func NewUserController(db *gorm.DB, userService *services.UserService) *UserController {
	return &UserController{
		db:          db,
		userService: userService,
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

	summary, err := uc.userService.ImportUsers(c.Request.Context(), openedFile)
	if err != nil {
		// Pass the error from the service to the error handling middleware
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "User import process completed.",
		"succeeded": summary.Succeeded,
		"failed":    summary.Failed,
		"failures":  summary.Failures,
	})
}

// GetUserAssets retrieves all assets owned by or shared with a specific user.
func (uc *UserController) GetUserAssets(c *gin.Context) {
	// Use the utility function to get the target user's ID from the URL param.
	targetUserID, err := utils.GetUUIDFromParam(c, "userId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Use the utility function to get the authenticated user's ID from the context.
	authUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
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