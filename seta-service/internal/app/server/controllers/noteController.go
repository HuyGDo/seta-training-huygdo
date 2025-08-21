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

// NoteController no longer embeds BaseController.
type NoteController struct {
	db *gorm.DB
}

// NewNoteController creates a new NoteController, injecting the db dependency.
func NewNoteController(db *gorm.DB) *NoteController {
	return &NoteController{db: db}
}

// GetNote retrieves a single note. Simplified with utils and auth middleware.
func (nc *NoteController) GetNote(c *gin.Context) {
	noteID, err := utils.GetUUIDFromParam(c, "noteId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	ctx := c.Request.Context()
	cacheKey := "note:" + noteID.String()
	var note models.Note

	// 1. Check Cache First (Cache-Aside)
	cachedNote, err := database.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit: Unmarshal and return
		if json.Unmarshal([]byte(cachedNote), &note) == nil {
			c.JSON(http.StatusOK, note)
			return
		}
	}

	// Cache Miss: Get from DB
	if err := nc.db.WithContext(ctx).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	// 2. Populate Cache
	noteJSON, _ := json.Marshal(note)
	database.Rdb.Set(ctx, cacheKey, noteJSON, 24*time.Hour)

	c.JSON(http.StatusOK, note)
}

type UpdateNoteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// UpdateNote updates a note's title or body. Simplified with utils and auth middleware.
func (nc *NoteController) UpdateNote(c *gin.Context) {
	noteID, err := utils.GetUUIDFromParam(c, "noteId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
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

	// Write-Through: Update the cache
	ctx := c.Request.Context()
	cacheKey := "note:" + note.NoteID.String()
	noteJSON, _ := json.Marshal(note)
	database.Rdb.Set(ctx, cacheKey, noteJSON, 24*time.Hour)

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType: "NOTE_UPDATED",
		AssetType: "note",
		AssetID:   note.NoteID.String(),
		OwnerID:   note.OwnerID.String(),
		ActionBy:  actorUserID.String(),
	})

	c.JSON(http.StatusOK, note)
}

// DeleteNote deletes a note. Simplified with utils and auth middleware.
func (nc *NoteController) DeleteNote(c *gin.Context) {
	noteID, err := utils.GetUUIDFromParam(c, "noteId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var note models.Note
	if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
		return
	}

	tx := nc.db.WithContext(c.Request.Context()).Begin()
	// ... (transaction logic remains the same)
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

	// Cache Invalidation (Write-Through for deletes)
	ctx := c.Request.Context()
	cacheKey := "note:" + noteID.String()
	database.Rdb.Del(ctx, cacheKey)

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType: "NOTE_DELETED",
		AssetType: "note",
		AssetID:   note.NoteID.String(),
		OwnerID:   note.OwnerID.String(),
		ActionBy:  actorUserID.String(),
	})

	c.Status(http.StatusNoContent)
}

type ShareNoteInput struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required"`
}

// ShareNote shares a note with another user. Simplified with utils and auth middleware.
func (nc *NoteController) ShareNote(c *gin.Context) {
	noteID, err := utils.GetUUIDFromParam(c, "noteId")
	if err != nil {
		_ = c.Error(err)
		return
	}

	actorUserID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var note models.Note
    if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
        return
    }

	var input ShareNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	share := models.NoteShare{
		NoteID: noteID,
		UserID: input.UserID,
		Access: input.Access,
	}

	if err := nc.db.WithContext(c.Request.Context()).Create(&share).Error; err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to share note"})
		return
	}

	// Write-Through: Immediately write to cache after DB success
	ctx := c.Request.Context()
	cacheKey := "note:" + note.NoteID.String()
	noteJSON, _ := json.Marshal(note)
	database.Rdb.Set(ctx, cacheKey, noteJSON, 24*time.Hour)

	go kafka.ProduceAssetEvent(context.Background(), kafka.EventPayload{
		EventType:    "NOTE_SHARED",
		AssetType:    "note",
		AssetID:      noteID.String(),
		OwnerID:      note.OwnerID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: input.UserID.String(),
	})

	c.Status(http.StatusNoContent)
}

// RevokeNoteSharing removes a user's access to a shared note. Simplified.
func (nc *NoteController) RevokeNoteSharing(c *gin.Context) {
	noteID, err := utils.GetUUIDFromParam(c, "noteId")
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

	var note models.Note
    if err := nc.db.WithContext(c.Request.Context()).First(&note, "note_id = ?", noteID).Error; err != nil {
        _ = c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: "Note not found"})
        return
    }

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
		AssetID:      noteID.String(),
		OwnerID:      note.OwnerID.String(),
		ActionBy:     actorUserID.String(),
		TargetUserID: targetUserID.String(),
	})

	c.Status(http.StatusNoContent)
}