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

type NoteController struct {
	BaseController
}

func NewNoteController(db *gorm.DB) *NoteController {
	return &NoteController{BaseController: NewBaseController(db)}
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

    actorUserID, ok := nc.GetUserIDFromContext(c)
    if !ok {
        return
    }

    // Authorization check: User must own the note or have direct/indirect write access.
    if note.OwnerID != actorUserID {
        var noteShare models.NoteShare
        err := nc.db.Where("note_id = ? AND user_id = ? AND access = 'write'", note.NoteID, actorUserID).First(&noteShare).Error
        if err != nil { // If no direct write access to the note, check for folder write access
            var folderShare models.FolderShare
            if err := nc.db.Where("folder_id = ? AND user_id = ? AND access = 'write'", note.FolderID, actorUserID).First(&folderShare).Error; err != nil {
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType:   "NOTE_UPDATED",
		AssetType:   "note",
		AssetID:     note.NoteID.String(),
		OwnerID:     note.OwnerID.String(),
		ActionBy: actorUserID.String(),
	})

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

    actorUserID, ok := nc.GetUserIDFromContext(c)
    if !ok {
        return
    }

    // Only the owner can delete a note
    if note.OwnerID != actorUserID {
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType:   "NOTE_DELETED",
		AssetType:   "note",
		AssetID:     note.NoteID.String(),
		OwnerID:     note.OwnerID.String(),
		ActionBy: actorUserID.String(),
	})

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

    actorUserID, ok := nc.GetUserIDFromContext(c)
    if !ok {
        return
    }

    // Only the owner can share a note
    if note.OwnerID != actorUserID {
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

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:    "NOTE_SHARED",
        AssetType:    "note",
        AssetID:      note.NoteID.String(),
        OwnerID:      note.OwnerID.String(),
        ActionBy:     actorUserID.String(),
        TargetUserID: input.UserID.String(),
    })

    c.Status(http.StatusNoContent)
}

// RevokeNoteSharing removes a user's access to a shared note.
func (nc *NoteController) RevokeNoteSharing(c *gin.Context) {
    noteID, err := uuid.Parse(c.Param("noteId"))
    if err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid note ID format"})
        return
    }

    targetUserID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "Invalid user ID format"})
        return
    }

    var note models.Note
    if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
        return
    }

    // Get the ID of the user performing the action from the context.
    actorUserID, ok := nc.GetUserIDFromContext(c)
    if !ok {
        return // Helper handles the error response.
    }

    // Authorization check: Only the owner of the note can revoke sharing.
    if note.OwnerID != actorUserID {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized to modify sharing for this note"})
        return
    }

    // Delete the specific share record.
    result := nc.db.WithContext(c.Request.Context()).
        Where("note_id = ? AND user_id = ?", noteID, targetUserID).
        Delete(&models.NoteShare{})

    if result.Error != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to revoke note share"})
        return
    }

    if result.RowsAffected == 0 {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Sharing record not found for this user and note"})
        return
    }

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
        EventType:    "NOTE_UNSHARED",
        AssetType:    "note",
        AssetID:      note.NoteID.String(),
        OwnerID:      note.OwnerID.String(),
        ActionBy:     actorUserID.String(),
        TargetUserID: targetUserID.String(),
    })

    c.Status(http.StatusNoContent)
}