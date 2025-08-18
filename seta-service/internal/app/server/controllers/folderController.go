package controllers

import (
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FolderController handles folder-related requests.
type FolderController struct {
	BaseController
}

// NewFolderController creates a new FolderController.
func NewFolderController(db *gorm.DB) *FolderController {
	return &FolderController{
		BaseController: NewBaseController(db),
	}
}

type CreateFolderInput struct {
	Name string `json:"name" binding:"required"`
}

func (fc *FolderController) CreateFolder(c *gin.Context) {
	var input CreateFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	// Use the helper from BaseController
	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response
	}

	folder := models.Folder{
		Name:    input.Name,
		OwnerID: userID,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&folder).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves a single folder
func (fc *FolderController) GetFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Authorization Check
	if folder.OwnerID != userID {
		var share models.FolderShare
		if err := fc.db.WithContext(c.Request.Context()).Where("folder_id = ? AND user_id = ?", folder.FolderID, userID).First(&share).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to view this folder"})
			return
		}
	}

	c.JSON(http.StatusOK, folder)
}

type UpdateFolderInput struct {
	Name string `json:"name" binding:"required"`
}

// UpdateFolder updates a folder's name
func (fc *FolderController) UpdateFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Authorization check
	if folder.OwnerID != userID {
		var share models.FolderShare
		if err := fc.db.WithContext(c.Request.Context()).Where("folder_id = ? AND user_id = ? AND access = 'write'", folder.FolderID, userID).First(&share).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to update this folder"})
			return
		}
	}

	var input UpdateFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := fc.db.WithContext(c.Request.Context()).Model(&folder).Update("name", input.Name).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder
func (fc *FolderController) DeleteFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	if folder.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to delete this folder"})
		return
	}

	// Use a transaction for safe deletion of folder and its associations
	tx := fc.db.WithContext(c.Request.Context()).Begin()
	if tx.Error != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to start transaction"})
		return
	}

	if err := tx.Where("folder_id = ?", folder.FolderID).Delete(&models.Note{}).Error; err != nil {
		tx.Rollback()
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete associated notes"})
		return
	}
	if err := tx.Where("folder_id = ?", folder.FolderID).Delete(&models.FolderShare{}).Error; err != nil {
		tx.Rollback()
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete associated shares"})
		return
	}
	if err := tx.Delete(&folder).Error; err != nil {
		tx.Rollback()
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete folder"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to commit transaction"})
		return
	}

	c.Status(http.StatusNoContent)
}

type ShareFolderInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"` // "read" or "write"
}

// ShareFolder shares a folder with another user
func (fc *FolderController) ShareFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Only the owner can share a folder
	if folder.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to share this folder"})
		return
	}

	var input ShareFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	share := models.FolderShare{
		FolderID: folder.FolderID,
		UserID:   input.UserID,
		Access:   input.Access,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share folder"})
		return
	}

	c.Status(http.StatusNoContent)
}