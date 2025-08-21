package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"seta/internal/pkg/database"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/kafka"
	"seta/internal/pkg/models"
	"seta/internal/pkg/utils" // Import the new utils package
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// It now holds its own database connection.
type FolderController struct {
	db *gorm.DB
}

// NewFolderController creates a new FolderController, injecting the db dependency.
func NewFolderController(db *gorm.DB) *FolderController {
	return &FolderController{
		db: db,
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

	// Use the new utility function to get the user ID from the context.
	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	folder := models.Folder{
		Name:    input.Name,
		OwnerID: userID,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&folder).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create folder"})
		return
	}
	
	// Cache Write through
	ctx := c.Request.Context()
	cacheKey := "folder:" + folder.FolderID.String()
	folderJSON, _ := json.Marshal(folder)
	database.Rdb.Set(ctx, cacheKey, folderJSON, 24*time.Hour)

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType: "FOLDER_CREATED",
		AssetType: "folder",
		AssetID:   folder.FolderID.String(),
		OwnerID:   folder.OwnerID.String(),
		ActionBy:  userID.String(),
	})

	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves a single folder
func (fc *FolderController) GetFolder(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	cacheKey := "folder:" + folderID.String()
	var folder models.Folder

	// Cache-Aside: Check cache first
	cachedFolder, err := database.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit
		json.Unmarshal([]byte(cachedFolder), &folder)
		c.JSON(http.StatusOK, folder)
		return
	}

	// Cache Miss: Get from DB
	if err := fc.db.WithContext(ctx).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	// Populate cache
	folderJSON, _ := json.Marshal(folder)
	database.Rdb.Set(ctx, cacheKey, folderJSON, 24*time.Hour)

	c.JSON(http.StatusOK, folder)
}

type UpdateFolderInput struct {
	Name string `json:"name" binding:"required"`
}

// UpdateFolder updates a folder's name
func (fc *FolderController) UpdateFolder(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
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

	// Write-Through: Update the cache
	ctx := c.Request.Context()
	cacheKey := "folder:" + folder.FolderID.String()
	folderJSON, _ := json.Marshal(folder)
	database.Rdb.Set(ctx, cacheKey, folderJSON, 24*time.Hour)

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType: "FOLDER_UPDATED",
		AssetType: "folder",
		AssetID:   folderID.String(),
		OwnerID:   userID.String(),
		ActionBy:  userID.String(),
	})

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder. Simplified with utils and auth middleware.
func (fc *FolderController) DeleteFolder(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var folder models.Folder
	if err := fc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	tx := fc.db.WithContext(c.Request.Context()).Begin()
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
		EventType: "FOLDER_DELETED",
		AssetType: "folder",
		AssetID:   folderID.String(),
		OwnerID:   folder.OwnerID.String(),
		ActionBy:  actorUserID.String(),
	})

	c.Status(http.StatusNoContent)
}

type ShareFolderInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"`
}

// ShareFolder shares a folder
func (fc *FolderController) ShareFolder(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var input ShareFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	share := models.FolderShare{
		FolderID: folderID,
		UserID:   input.UserID,
		Access:   input.Access,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share folder"})
		return
	}

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType:    "FOLDER_SHARED",
		AssetType:    "folder",
		AssetID:      folderID.String(),
		OwnerID:      actorUserID.String(), // The actor is the owner
		ActionBy:     actorUserID.String(),
		TargetUserID: input.UserID.String(),
	})

	c.Status(http.StatusNoContent)
}

// RevokeFolderSharing removes a user's access
func (fc *FolderController) RevokeFolderSharing(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	targetUserID, err := utils.GetUUIDFromParam(c, "userId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	
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
		AssetID:      folderID.String(),
		OwnerID:      actorUserID.String(), // The actor is the owner
		ActionBy:     actorUserID.String(),
		TargetUserID: targetUserID.String(),
	})

	c.Status(http.StatusNoContent)
}

type CreateNoteInput struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"`
}

// CreateNote creates a new note inside a folder
func (fc *FolderController) CreateNote(c *gin.Context) {
	folderID, err := utils.GetUUIDFromParam(c, "folderId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	note := models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	if err := fc.db.WithContext(c.Request.Context()).Create(&note).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create note"})
		return
	}

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType: "NOTE_CREATED",
		AssetType: "note",
		AssetID:   note.NoteID.String(),
		OwnerID:   note.OwnerID.String(),
		ActionBy:  userID.String(),
	})

	c.JSON(http.StatusCreated, note)
}