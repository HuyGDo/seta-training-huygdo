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
	teams.Use(middlewares.IsAuthorizedRole("MANAGER"))
	{
		teams.POST("", teamController.CreateTeam)
		teams.POST("/:teamId/members", middlewares.IsTeamManager(db), teamController.AddMember)
		teams.DELETE("/:teamId/members/:memberId", middlewares.IsTeamManager(db), teamController.RemoveMember)
		teams.POST("/:teamId/managers", middlewares.IsLeadManager(db), teamController.AddManager)
		teams.DELETE("/:teamId/managers/:managerId", middlewares.IsLeadManager(db), teamController.RemoveManager)
		teams.GET("/:teamId/assets", middlewares.IsTeamManager(db), teamController.GetTeamAssets)
	}
}