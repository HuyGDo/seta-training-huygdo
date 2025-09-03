package bootstrap

import (
	http "seta/internal/adapters/http/handlers"
	"seta/internal/adapters/http/middleware"
	"seta/internal/application/ports"
	"seta/pkg/errorHandling"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func SetupRouter(
	log *zerolog.Logger,
	authService ports.AuthService,
	teamRepo ports.TeamRepository, // Needed for middleware
	teamHandler *http.TeamHandler,
	folderHandler *http.FolderHandler,
	noteHandler *http.NoteHandler,
	userHandler *http.UserHandler,
) *gin.Engine {
	r := gin.Default()

	// Global Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.PrometheusMiddleware())
	r.Use(errorHandling.ErrorHandler())

	// Public Routes
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API Group with Authentication
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authService))
	{
		// Team Routes
		teams := api.Group("/teams")
		teams.Use(middleware.IsAuthorizedRole("MANAGER"))
		{
			teams.POST("", teamHandler.CreateTeam)
			teams.POST("/:teamId/members", middleware.IsTeamManagerMiddleware(teamRepo), teamHandler.AddMember)
			teams.DELETE("/:teamId/members/:memberId", middleware.IsTeamManagerMiddleware(teamRepo), teamHandler.RemoveMember)
			teams.POST("/:teamId/managers", middleware.IsLeadManagerMiddleware(teamRepo), teamHandler.AddManager)
			teams.DELETE("/:teamId/managers/:managerId", middleware.IsLeadManagerMiddleware(teamRepo), teamHandler.RemoveManager)
			teams.GET("/:teamId/assets", middleware.IsTeamManagerMiddleware(teamRepo), teamHandler.GetTeamAssets)
		}

		// User Routes
		users := api.Group("/users")
		{
			// Auth middleware already confirms user is valid. Further checks are in the use case.
			users.GET("/:userId/assets", userHandler.GetUserAssets)
			users.POST("/import", middleware.IsAuthorizedRole("MANAGER"), userHandler.ImportUsers)
		}

		// Folder Routes
		folders := api.Group("/folders")
		{
			folders.POST("", folderHandler.CreateFolder)
			// Middleware for specific folder access would be added here if needed
			// For now, authorization is handled within the use cases
			folders.PUT("/:folderId", folderHandler.UpdateFolder)
			folders.POST("/:folderId/notes", folderHandler.CreateNoteInFolder)
			folders.POST("/:folderId/share", folderHandler.ShareFolder)
			folders.DELETE("/:folderId/share/:userId", folderHandler.RevokeFolderSharing)
		}

		// Note Routes
		notes := api.Group("/notes")
		{
			notes.GET("/:noteId", noteHandler.GetNote)
			notes.PUT("/:noteId", noteHandler.UpdateNote)
			notes.DELETE("/:noteId", noteHandler.DeleteNote)
			notes.POST("/:noteId/share", noteHandler.ShareNote)
			notes.DELETE("/:noteId/share/:userId", noteHandler.RevokeNoteSharing)
		}
	}

	return r
}
