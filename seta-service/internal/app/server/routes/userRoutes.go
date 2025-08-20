package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/middlewares"
	"seta/internal/app/server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterUserRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	userService := services.NewUserService()
	userController := controllers.NewUserController(db, userService)

	users := rg.Group("/users")
	{
		users.GET("/:userId/assets", userController.GetUserAssets)
		users.POST("/import", middlewares.IsAuthorizedRole("MANAGER"), userController.ImportUsers)
	}
}