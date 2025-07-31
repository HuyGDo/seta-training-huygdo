package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/graphql"
	"seta/internal/app/server/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SetupRouter initializes the Gin router and sets up the routes.
func SetupRouter(db *gorm.DB, log *logrus.Logger) *gin.Engine {
	r := gin.Default()

	// Add Prometheus middleware to all routes
	r.Use(middlewares.PrometheusMiddleware())
	r.Use(middlewares.ErrorHandler())

	// Add Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// GraphQL endpoint
	r.POST("/graphql", graphql.GraphQLHandler(db, log))
	r.GET("/", graphql.PlaygroundHandler())

	// REST endpoints for team and asset management
	api := r.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		// Team routes
		teamController := controllers.NewTeamController(db)
		teams := api.Group("/teams")
		teams.Use(middlewares.IsAuthorized("manager"))
		{
			teams.POST("", teamController.CreateTeam)
			teams.POST("/:teamId/members", teamController.AddMember)
			teams.DELETE("/:teamId/members/:memberId", teamController.RemoveMember)
			teams.POST("/:teamId/managers", teamController.AddManager)
			teams.DELETE("/:teamId/managers/:managerId", teamController.RemoveManager)
			teams.GET("/:teamId/assets", middlewares.IsTeamManager(db), teamController.GetTeamAssets)
		}

		assetController := controllers.NewAssetController(db)

		// User routes
		users := api.Group("/users")
		{
			users.GET("/:userId/assets", assetController.GetUserAssets)
		}

		// Asset routes
		assets := api.Group("/assets")
		{
			// Folder routes
			folders := assets.Group("/folders")
			{
				folders.POST("", assetController.CreateFolder)
				folders.GET("/:folderId", assetController.GetFolder)
				folders.PUT("/:folderId", assetController.UpdateFolder)
				folders.DELETE("/:folderId", assetController.DeleteFolder)
				folders.POST("/:folderId/share", assetController.ShareFolder)
			}

			// Note routes
			notes := assets.Group("/notes")
			{
				notes.POST("", assetController.CreateNote)
				notes.GET("/:noteId", assetController.GetNote)
				notes.PUT("/:noteId", assetController.UpdateNote)
				notes.DELETE("/:noteId", assetController.DeleteNote)
				notes.POST("/:noteId/share", assetController.ShareNote)
			}
		}
	}

	return r
}
