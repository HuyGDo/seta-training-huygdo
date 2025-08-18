package controllers

import (
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NoteController struct {
	BaseController
}

func NewNoteController(db *gorm.DB) *NoteController {
	return &NoteController{BaseController: NewBaseController(db)}
}

type CreateNoteInput struct {
	Title    string    `json:"title" binding:"required"`
	Body     string    `json:"body"`
	FolderID uuid.UUID `json:"folderId" binding:"required"`
}

// CreateNote creates a new note
func (nc *NoteController) CreateNote(c *gin.Context) {
	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userID, ok := nc.GetUserIDFromContext(c)
	if !ok {
		return // Helper handles the error response
	}

	// Verify that the user has write access to the folder
	var folder models.Folder
	if err := nc.db.WithContext(c.Request.Context()).First(&folder, "folder_id = ?", input.FolderID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Folder not found"})
		return
	}
	if folder.OwnerID != userID {
		var share models.FolderShare
		if err := nc.db.WithContext(c.Request.Context()).Where("folder_id = ? AND user_id = ? AND access = 'write'", folder.FolderID, userID).First(&share).Error; err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to create a note in this folder"})
			return
		}
	}

	note := models.Note{
		Title:    input.Title,
		Body:     input.Body,
		FolderID: input.FolderID,
		OwnerID:  userID,
	}

	if err := nc.db.WithContext(c.Request.Context()).Create(&note).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote retrieves a single note
func (nc *NoteController) GetNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userID, ok := nc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Authorization check: User must own the note, have direct access to the note, or have access to its parent folder.
	if note.OwnerID != userID {
		var noteShare models.NoteShare
		err := nc.db.Where("note_id = ? AND user_id = ?", note.NoteID, userID).First(&noteShare).Error
		if err != nil { // If no direct note share, check for folder share
			var folderShare models.FolderShare
			if err := nc.db.Where("folder_id = ? AND user_id = ?", note.FolderID, userID).First(&folderShare).Error; err != nil {
				_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to view this note"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, note)
}

type UpdateNoteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// UpdateNote updates a note's title or body
func (nc *NoteController) UpdateNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userID, ok := nc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Authorization check: User must own the note or have direct/indirect write access.
	if note.OwnerID != userID {
		var noteShare models.NoteShare
		err := nc.db.Where("note_id = ? AND user_id = ? AND access = 'write'", note.NoteID, userID).First(&noteShare).Error
		if err != nil { // If no direct write access to the note, check for folder write access
			var folderShare models.FolderShare
			if err := nc.db.Where("folder_id = ? AND user_id = ? AND access = 'write'", note.FolderID, userID).First(&folderShare).Error; err != nil {
				_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to update this note"})
				return
			}
		}
	}

	var input UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if err := nc.db.WithContext(c.Request.Context()).Model(&note).Updates(input).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note
func (nc *NoteController) DeleteNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userID, ok := nc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Only the owner can delete a note
	if note.OwnerID != userID {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to delete this note"})
		return
	}
	
	tx := nc.db.WithContext(c.Request.Context()).Begin()
    if tx.Error != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to start transaction"})
        return
    }

	if err := tx.Where("note_id = ?", note.NoteID).Delete(&models.NoteShare{}).Error; err != nil {
        tx.Rollback()
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete associated shares"})
        return
    }
	if err := tx.Delete(&note).Error; err != nil {
        tx.Rollback()
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to delete note"})
		return
	}

    if err := tx.Commit().Error; err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to commit transaction"})
        return
    }

	c.Status(http.StatusNoContent)
}

type ShareNoteInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"` // "read" or "write"
}

// ShareNote shares a note with another user
func (nc *NoteController) ShareNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	userID, ok := nc.GetUserIDFromContext(c)
	if !ok {
		return
	}

	// Only the owner can share a note
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

	if err := nc.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share note"})
		return
	}

	c.Status(http.StatusNoContent)
}