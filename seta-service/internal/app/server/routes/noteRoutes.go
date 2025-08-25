package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterNoteRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	noteController := controllers.NewNoteController(db)
	notes := rg.Group("/notes")
	{
		// Note creation is now under folder routes.
		notes.GET("/:noteId", middlewares.CanReadNote(db), noteController.GetNote)
		notes.PUT("/:noteId", middlewares.CanWriteNote(db), noteController.UpdateNote)
		notes.DELETE("/:noteId", middlewares.IsNoteOwner(db), noteController.DeleteNote)
		notes.POST("/:noteId/share", middlewares.IsNoteOwner(db), noteController.ShareNote)
		notes.DELETE("/:noteId/share/:userId", middlewares.IsNoteOwner(db), noteController.RevokeNoteSharing)
	}
}