package controllers

import (
	"context"
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/kafka"
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:   "FOLDER_CREATED",
        AssetType:   "folder",
        AssetID:     folder.FolderID.String(),
        OwnerID:     folder.OwnerID.String(),
        ActionBy:    userID.String(),
    })

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

// Maybe N+1 query problem --> will check later
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:   "FOLDER_UPDATED",
        AssetType:   "folder",
        AssetID:     folderID.String(),
        OwnerID:     userID.String(),
        ActionBy: userID.String(), // userID from GetUserIDFromContext
    })

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

	actorUserID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	if folder.OwnerID != actorUserID {
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:   "FOLDER_DELETED",
        AssetType:   "folder",
        AssetID:     folderID.String(),
        OwnerID:     folder.OwnerID.String(),
        ActionBy: actorUserID.String(), // userID from GetUserIDFromContext
    })

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

	actorUserID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Only the owner can share a folder
	if folder.OwnerID != actorUserID {
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

	targetUserID := input.UserID
	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:    "FOLDER_SHARED",
        AssetType:    "folder",
        AssetID:      folder.FolderID.String(),
        OwnerID:      folder.OwnerID.String(),
        ActionBy:     actorUserID.String(),
        TargetUserID: targetUserID.String(),
    })

	c.Status(http.StatusNoContent)
}

// RevokeFolderSharing removes a user's access to a shared folder.
func (fc *FolderController) RevokeFolderSharing(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid user ID format"})
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	// Get the ID of the user performing the action from the context.
	actorUserID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response.
	}

	// Authorization check: Only the owner of the folder can revoke sharing.
	if folder.OwnerID != actorUserID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to modify sharing for this folder"})
		return
	}

	// Delete the specific share record.
	result := fc.db.WithContext(c.Request.Context()).
		Where("folder_id = ? AND user_id = ?", folderID, targetUserID).
		Delete(&models.FolderShare{})

	if result.Error != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to revoke folder share"})
		return
	}

	if result.RowsAffected == 0 {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Sharing record not found for this user and folder"})
		return
	}

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:    "FOLDER_UNSHARED",
        AssetType:    "folder",
        AssetID:      folder.FolderID.String(),
        OwnerID:      folder.OwnerID.String(),
        ActionBy:     actorUserID.String(),
        TargetUserID: targetUserID.String(),
    })

	c.Status(http.StatusNoContent)
}

type CreateNoteInput struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"`
}

// CreateNote creates a new note
func (fc *FolderController) CreateNote(c *gin.Context) {
	// Get folderId from the URL parameter
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userID, ok := fc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response
	}

	// Verify that the user has write access to the folder
	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}
	if folder.OwnerID != userID {
		var share models.FolderShare
		if err := fc.db.WithContext(c.Request.Context()).Where("folder_id = ? AND user_id = ? AND access = 'write'", folder.FolderID, userID).First(&share).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to create a note in this folder"})
			return
		}
	}

	note := models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: folderID, // Use folderID from the URL
		OwnerID:  userID,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&note).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create note"})
		return
	}

	// Corrected Kafka Event for NOTE_CREATED
	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType:   "NOTE_CREATED",
		AssetType:   "note",
		AssetID:     note.NoteID.String(),
		OwnerID:     note.OwnerID.String(),
		ActionBy: userID.String(),
	})

	c.JSON(http.StatusCreated, note)
}