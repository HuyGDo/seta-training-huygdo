package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterTeamRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	teamController := controllers.NewTeamController(db)
	teams := rg.Group("/teams")
	teams.Use(middlewares.IsAuthorized("MANAGER"))
	{
		teams.POST("", teamController.CreateTeam)
		teams.POST("/:teamId/members", teamController.AddMember)
		teams.DELETE("/:teamId/members/:memberId", teamController.RemoveMember)
		teams.POST("/:teamId/managers", teamController.AddManager)
		teams.DELETE("/:teamId/managers/:managerId", teamController.RemoveManager)
		teams.GET("/:teamId/assets", middlewares.IsTeamManager(db), teamController.GetTeamAssets)
	}
}