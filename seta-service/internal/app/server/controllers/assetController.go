package controllers

import (
	"net/http"

	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetController handles asset-related requests.
type AssetController struct {
	db *gorm.DB
}

// NewAssetController creates a new AssetController.
func NewAssetController(db *gorm.DB) *AssetController {
	return &AssetController{db: db}
}

// GetTeamAssets retrieves assets for a team.
func (ac *AssetController) GetTeamAssets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetTeamAssets placeholder"})
}

type CreateFolderInput struct {
	Name string `json:"name" binding:"required"`
}

// CreateFolder creates a new folder
func (ac *AssetController) CreateFolder(c *gin.Context) {
	var input CreateFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// set a constructed error response
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	folder := models.Folder{
		Name:    input.Name,
		OwnerID: userID,
	}

	if err := ac.db.WithContext(c.Request.Context()).Create(&folder).Error; err != nil {
		// log.Error...
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves a single folder
func (ac *AssetController) GetFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := ac.db.WithContext(c.Request.Context()).First(&folder, folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if folder.OwnerID != userID {
		// Check if the folder is shared with the user
		var share models.FolderShare
		// don't recommend this, it is hardcode, when you change column, you need to change this too
		// try to use like this:
		//db.Where(&model.FolderShare{FolderID: "abc", UserID: 20})
		if err := ac.db.WithContext(c.Request.Context()).Where("folder_id = ? AND user_id = ?", folder.FolderID, userID).First(&share).Error; err != nil {
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
func (ac *AssetController) UpdateFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := ac.db.WithContext(c.Request.Context()).First(&folder, folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if folder.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to update this folder"})
		return
	}

	var input UpdateFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := ac.db.WithContext(c.Request.Context()).Model(&folder).Update("name", input.Name).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder
func (ac *AssetController) DeleteFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := ac.db.WithContext(c.Request.Context()).First(&folder, folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if folder.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to delete this folder"})
		return
	}

	// Delete associated notes and shares
	// handle error
	ac.db.WithContext(c.Request.Context()).Where("folder_id = ?", folder.FolderID).Delete(&models.Note{})
	ac.db.WithContext(c.Request.Context()).Where("folder_id = ?", folder.FolderID).Delete(&models.FolderShare{})
	if err := ac.db.WithContext(c.Request.Context()).Delete(&folder).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete folder"})
		return
	}

	c.Status(http.StatusNoContent)
}

type ShareFolderInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"` // "read" or "write"
}

// ShareFolder shares a folder with another user
func (ac *AssetController) ShareFolder(c *gin.Context) {
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid folder ID format"})
		return
	}

	var folder models.Folder
	if err := ac.db.WithContext(c.Request.Context()).First(&folder, folderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

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

	if err := ac.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share folder"})
		return
	}

	c.Status(http.StatusNoContent)
}

type CreateNoteInput struct {
	Title    string    `json:"title" binding:"required"`
	Body     string    `json:"body"`
	FolderID uuid.UUID `json:"folderId" binding:"required"`
}

// CreateNote creates a new note
func (ac *AssetController) CreateNote(c *gin.Context) {
	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	// Verify that the user has access to the folder
	var folder models.Folder
	if err := ac.db.WithContext(c.Request.Context()).First(&folder, input.FolderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}
	if folder.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to create a note in this folder"})
		return
	}

	note := models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: input.FolderID,
		OwnerID:  userID,
	}

	if err := ac.db.WithContext(c.Request.Context()).Create(&note).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote retrieves a single note
func (ac *AssetController) GetNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := ac.db.WithContext(c.Request.Context()).First(&note, noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if note.OwnerID != userID {
		// Check if the note is shared with the user
		var share models.NoteShare
		if err := ac.db.WithContext(c.Request.Context()).Where("note_id = ? AND user_id = ?", note.NoteID, userID).First(&share).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to view this note"})
			return
		}
	}

	c.JSON(http.StatusOK, note)
}

type UpdateNoteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// UpdateNote updates a note's title or body
func (ac *AssetController) UpdateNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := ac.db.WithContext(c.Request.Context()).First(&note, noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if note.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to update this note"})
		return
	}

	var input UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := ac.db.WithContext(c.Request.Context()).Model(&note).Updates(input).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note
func (ac *AssetController) DeleteNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := ac.db.WithContext(c.Request.Context()).First(&note, noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if note.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to delete this note"})
		return
	}

	// Delete associated shares
	ac.db.WithContext(c.Request.Context()).Where("note_id = ?", note.NoteID).Delete(&models.NoteShare{})
	if err := ac.db.WithContext(c.Request.Context()).Delete(&note).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete note"})
		return
	}

	c.Status(http.StatusNoContent)
}

type ShareNoteInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"` // "read" or "write"
}

// ShareNote shares a note with another user
func (ac *AssetController) ShareNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := ac.db.WithContext(c.Request.Context()).First(&note, noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if note.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to share this note"})
		return
	}

	var input ShareNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	share := models.NoteShare{
		NoteID: note.NoteID,
		UserID: input.UserID,
		Access: input.Access,
	}

	if err := ac.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share note"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserAssets retrieves all assets owned by or shared with a specific user.
func (ac *AssetController) GetUserAssets(c *gin.Context) {
	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid user ID"})
		return
	}

	authUserIDStr, exists := c.Get("userId")
	if !exists {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "User not authenticated"})
		return
	}

	authUserID, err := uuid.Parse(authUserIDStr.(string))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Invalid user ID format"})
		return
	}

	if authUserID != targetUserID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to view these assets"})
		return
	}

	var folders []models.Folder
	if err := ac.db.WithContext(c.Request.Context()).
		Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id = ? OR folder_shares.user_id = ?", targetUserID, targetUserID).
		Group("folders.folder_id").
		Find(&folders).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to retrieve folders for the user"})
		return
	}

	var notes []models.Note
	if err := ac.db.WithContext(c.Request.Context()).
		Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("notes.owner_id = ? OR note_shares.user_id = ?", targetUserID, targetUserID).
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
