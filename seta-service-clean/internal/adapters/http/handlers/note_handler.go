package handlers

import (
	"net/http"
	"seta/internal/adapters/http/utils"
	"seta/internal/application"
	"seta/internal/domain/common"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	getNoteUseCase    *application.GetNoteUseCase
	updateNoteUseCase *application.UpdateNoteUseCase
	deleteNoteUseCase *application.DeleteNoteUseCase
	shareNoteUseCase  *application.ShareNoteUseCase
	unshareNoteUseCase *application.UnshareNoteUseCase
}

func NewNoteHandler(
	getUC *application.GetNoteUseCase, updateUC *application.UpdateNoteUseCase,
	deleteUC *application.DeleteNoteUseCase, shareUC *application.ShareNoteUseCase,
	unshareUC *application.UnshareNoteUseCase,
) *NoteHandler {
	return &NoteHandler{
		getNoteUseCase: getUC, updateNoteUseCase: updateUC,
		deleteNoteUseCase: deleteUC, shareNoteUseCase: shareUC,
		unshareNoteUseCase: unshareUC,
	}
}

// DTOs
type updateNoteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Handlers
func (h *NoteHandler) GetNote(c *gin.Context) {
	noteID, ok := utils.GetUUIDFromParam(c, "noteId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.GetNoteInput{
		NoteID: common.NoteID(noteID), RequesterID: requesterID,
	}
	note, err := h.getNoteUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusNotFound, Message: err.Error()}); return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) UpdateNote(c *gin.Context) {
	noteID, ok := utils.GetUUIDFromParam(c, "noteId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }
	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.UpdateNoteInput{
		NoteID: common.NoteID(noteID), Title: req.Title, Body: req.Body, RequesterID: requesterID,
	}
	note, err := h.updateNoteUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {
	noteID, ok := utils.GetUUIDFromParam(c, "noteId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.DeleteNoteInput{
		NoteID: common.NoteID(noteID), RequesterID: requesterID,
	}
	if err := h.deleteNoteUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *NoteHandler) ShareNote(c *gin.Context) {
	noteID, ok := utils.GetUUIDFromParam(c, "noteId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }
	var req shareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.ShareNoteInput{
		NoteID:      common.NoteID(noteID),
		ShareWithID: common.UserID(req.UserID),
		AccessLevel: common.Access(req.Access),
		RequesterID: requesterID,
	}
	if err := h.shareNoteUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *NoteHandler) RevokeNoteSharing(c *gin.Context) {
	noteID, ok := utils.GetUUIDFromParam(c, "noteId"); if !ok { return }
	targetUserID, ok := utils.GetUUIDFromParam(c, "userId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.UnshareNoteInput{
		NoteID: common.NoteID(noteID), UnshareWithID: common.UserID(targetUserID), RequesterID: requesterID,
	}
	if err := h.unshareNoteUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}
