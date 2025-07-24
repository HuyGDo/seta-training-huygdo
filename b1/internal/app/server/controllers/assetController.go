package controllers

import (
	"net/http"
	"strconv"

	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userId")
	folder := models.Folder{
		Name:    input.Name,
		OwnerID: userID.(uint),
	}

	if err := ac.db.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// GetFolder retrieves a single folder
func (ac *AssetController) GetFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var folder models.Folder

	if err := ac.db.First(&folder, folderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	userID, _ := c.Get("userId")
	if folder.OwnerID != userID.(uint) {
		// Check if the folder is shared with the user
		var share models.FolderShare
		if err := ac.db.Where("folder_id = ? AND user_id = ?", folder.FolderID, userID).First(&share).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this folder"})
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
	folderID := c.Param("folderId")
	var folder models.Folder

	if err := ac.db.First(&folder, folderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	userID, _ := c.Get("userId")
	if folder.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this folder"})
		return
	}

	var input UpdateFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.db.Model(&folder).Update("name", input.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

// DeleteFolder deletes a folder
func (ac *AssetController) DeleteFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var folder models.Folder

	if err := ac.db.First(&folder, folderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	userID, _ := c.Get("userId")
	if folder.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this folder"})
		return
	}

	// Delete associated notes and shares
	ac.db.Where("folder_id = ?", folder.FolderID).Delete(&models.Note{})
	ac.db.Where("folder_id = ?", folder.FolderID).Delete(&models.FolderShare{})
	if err := ac.db.Delete(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted successfully"})
}

type ShareFolderInput struct {
	UserID uint   `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required"` // "read" or "write"
}

// ShareFolder shares a folder with another user
func (ac *AssetController) ShareFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var folder models.Folder

	if err := ac.db.First(&folder, folderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	userID, _ := c.Get("userId")
	if folder.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to share this folder"})
		return
	}

	var input ShareFolderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share := models.FolderShare{
		FolderID: folder.FolderID,
		UserID:   input.UserID,
		Access:   input.Access,
	}

	if err := ac.db.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

type CreateNoteInput struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body"`
	FolderID uint   `json:"folderId" binding:"required"`
}

// CreateNote creates a new note
func (ac *AssetController) CreateNote(c *gin.Context) {
	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userId")

	// Verify that the user has access to the folder
	var folder models.Folder
	if err := ac.db.First(&folder, input.FolderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if folder.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to create a note in this folder"})
		return
	}

	note := models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: input.FolderID,
		OwnerID:  userID.(uint),
	}

	if err := ac.db.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote retrieves a single note
func (ac *AssetController) GetNote(c *gin.Context) {
	noteID := c.Param("noteId")
	var note models.Note

	if err := ac.db.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	userID, _ := c.Get("userId")
	if note.OwnerID != userID.(uint) {
		// Check if the note is shared with the user
		var share models.NoteShare
		if err := ac.db.Where("note_id = ? AND user_id = ?", note.NoteID, userID).First(&share).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this note"})
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
	noteID := c.Param("noteId")
	var note models.Note

	if err := ac.db.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	userID, _ := c.Get("userId")
	if note.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this note"})
		return
	}

	var input UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.db.Model(&note).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note
func (ac *AssetController) DeleteNote(c *gin.Context) {
	noteID := c.Param("noteId")
	var note models.Note

	if err := ac.db.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	userID, _ := c.Get("userId")
	if note.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this note"})
		return
	}

	// Delete associated shares
	ac.db.Where("note_id = ?", note.NoteID).Delete(&models.NoteShare{})
	if err := ac.db.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

type ShareNoteInput struct {
	UserID uint   `json:"userId" binding:"required"`
	Access string `json:"access" binding:"required"` // "read" or "write"
}

// ShareNote shares a note with another user
func (ac *AssetController) ShareNote(c *gin.Context) {
	noteID := c.Param("noteId")
	var note models.Note

	if err := ac.db.First(&note, noteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	userID, _ := c.Get("userId")
	if note.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to share this note"})
		return
	}

	var input ShareNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share := models.NoteShare{
		NoteID: note.NoteID,
		UserID: input.UserID,
		Access: input.Access,
	}

	if err := ac.db.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note shared successfully"})
}

// GetUserAssets retrieves all assets owned by or shared with a specific user.
func (ac *AssetController) GetUserAssets(c *gin.Context) {
	targetUserID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	authUserID, _ := c.Get("userId")
	if authUserID.(uint) != uint(targetUserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view these assets"})
		return
	}

	var folders []models.Folder
	if err := ac.db.
		Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id = ? OR folder_shares.user_id = ?", targetUserID, targetUserID).
		Group("folders.folder_id").
		Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve folders for the user"})
		return
	}

	var notes []models.Note
	if err := ac.db.
		Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("notes.owner_id = ? OR note_shares.user_id = ?", targetUserID, targetUserID).
		Group("notes.note_id").
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes for the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"notes":   notes,
	})
}
