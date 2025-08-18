package routes

import (
	"seta/internal/app/server/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterNoteRoutes(rg *gin.RouterGroup, db *gorm.DB) {
    noteController := controllers.NewNoteController(db)
    notes := rg.Group("/notes")
    {
        notes.POST("", noteController.CreateNote)
        notes.GET("/:noteId", noteController.GetNote)
        notes.PUT("/:noteId", noteController.UpdateNote)
        notes.DELETE("/:noteId", noteController.DeleteNote)
        notes.POST("/:noteId/share", noteController.ShareNote)
    }
}