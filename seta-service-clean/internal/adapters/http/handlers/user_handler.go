package handlers

import (
	"net/http"
	"seta/internal/application"
	"seta/internal/domain/common"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	importUsersUseCase   *application.ImportUsersUseCase
	getUserAssetsUseCase *application.GetUserAssetsUseCase
}

func NewUserHandler(importUC *application.ImportUsersUseCase, getAssetsUC *application.GetUserAssetsUseCase) *UserHandler {
	return &UserHandler{
		importUsersUseCase:   importUC,
		getUserAssetsUseCase: getAssetsUC,
	}
}

func (h *UserHandler) ImportUsers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "File not provided in 'file' form field"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to open uploaded file"})
		return
	}
	defer openedFile.Close()

	summary, err := h.importUsersUseCase.Execute(c.Request.Context(), openedFile)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *UserHandler) GetUserAssets(c *gin.Context) {
	targetUserID, _ := uuid.Parse(c.Param("userId"))
	requesterIDStr, _ := c.Get("userId")
	requesterID, _ := uuid.Parse(requesterIDStr.(string))

	input := application.GetUserAssetsInput{
		TargetUserID: common.UserID(targetUserID),
		RequesterID:  common.UserID(requesterID),
	}

	assets, err := h.getUserAssetsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, assets)
}
