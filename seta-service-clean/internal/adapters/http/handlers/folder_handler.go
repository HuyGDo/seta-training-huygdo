package handlers

import (
	"net/http"
	"seta/internal/adapters/http/utils"
	"seta/internal/application"
	"seta/internal/domain/common"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FolderHandler struct {
	createFolderUseCase       *application.CreateFolderUseCase
	updateFolderUseCase       *application.UpdateFolderUseCase
	createNoteInFolderUseCase *application.CreateNoteInFolderUseCase
	shareFolderUseCase        *application.ShareFolderUseCase
	unshareFolderUseCase      *application.UnshareFolderUseCase
	// getFolder and deleteFolder use cases would also be injected here
}

func NewFolderHandler(
	createUC *application.CreateFolderUseCase, updateUC *application.UpdateFolderUseCase,
	createNoteUC *application.CreateNoteInFolderUseCase, shareUC *application.ShareFolderUseCase,
	unshareUC *application.UnshareFolderUseCase,
) *FolderHandler {
	return &FolderHandler{
		createFolderUseCase: createUC, updateFolderUseCase: updateUC,
		createNoteInFolderUseCase: createNoteUC, shareFolderUseCase: shareUC,
		unshareFolderUseCase: unshareUC,
	}
}

// DTOs
type createFolderRequest struct {
	Name string `json:"name" binding:"required"`
}
type updateFolderRequest struct {
	Name string `json:"name" binding:"required"`
}
type createNoteRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"`
}
type shareRequest struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Access string    `json:"access" binding:"required,oneof=read write"`
}

// Handlers
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var req createFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}
	ownerID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.CreateFolderInput{Name: req.Name, OwnerID: ownerID}
	folder, err := h.createFolderUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: err.Error()}); return
	}
	c.JSON(http.StatusCreated, folder)
}

func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	folderID, ok := utils.GetUUIDFromParam(c, "folderId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }
	var req updateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.UpdateFolderInput{
		FolderID: common.FolderID(folderID), Name: req.Name, RequesterID: requesterID,
	}
	folder, err := h.updateFolderUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.JSON(http.StatusOK, folder)
}

func (h *FolderHandler) CreateNoteInFolder(c *gin.Context) {
	folderID, ok := utils.GetUUIDFromParam(c, "folderId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.CreateNoteInFolderInput{
		Title: req.Title, Body: req.Body, FolderID: common.FolderID(folderID), RequesterID: requesterID,
	}
	note, err := h.createNoteInFolderUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.JSON(http.StatusCreated, note)
}

func (h *FolderHandler) ShareFolder(c *gin.Context) {
	folderID, ok := utils.GetUUIDFromParam(c, "folderId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }
	var req shareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()}); return
	}

	input := application.ShareFolderInput{
		FolderID:    common.FolderID(folderID),
		ShareWithID: common.UserID(req.UserID),
		AccessLevel: common.Access(req.Access),
		RequesterID: requesterID,
	}
	if err := h.shareFolderUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}

func (h *FolderHandler) RevokeFolderSharing(c *gin.Context) {
	folderID, ok := utils.GetUUIDFromParam(c, "folderId"); if !ok { return }
	targetUserID, ok := utils.GetUUIDFromParam(c, "userId"); if !ok { return }
	requesterID, ok := utils.GetUserIDFromContext(c); if !ok { return }

	input := application.UnshareFolderInput{
		FolderID: common.FolderID(folderID), UnshareWithID: common.UserID(targetUserID), RequesterID: requesterID,
	}
	if err := h.unshareFolderUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()}); return
	}
	c.Status(http.StatusNoContent)
}
