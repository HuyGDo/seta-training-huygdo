package controllers

import (
	"net/http"
	"seta/internal/app/server/services"
	"seta/internal/pkg/errorHandling"

	"github.com/gin-gonic/gin"
)

// UserController handles user-related HTTP requests.
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new UserController.
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// ImportUsers handles the file upload and calls the user service to process it.
func (uc *UserController) ImportUsers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: "File not provided in 'file' form field"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to open uploaded file"})
		return
	}
	defer openedFile.Close()

	summary, err := uc.userService.ImportUsers(openedFile)
	if err != nil {
		_ = c.Error(&errorHandling.CustomError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "User import process completed.",
		"succeeded": summary.Succeeded,
		"failed":    summary.Failed,
	})
}