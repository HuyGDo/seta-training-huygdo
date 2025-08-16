package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/middlewares"
	"seta/internal/app/server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterUserRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	assetController := controllers.NewAssetController(db)
	userService := services.NewUserService()
	userController := controllers.NewUserController(userService)

	users := rg.Group("/users")
	{
		users.GET("/:userId/assets", assetController.GetUserAssets)
		users.POST("/import", middlewares.IsAuthorized("MANAGER"), userController.ImportUsers)
	}
}