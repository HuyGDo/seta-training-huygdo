package routes

import (
	"seta/internal/app/server/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterNoteRoutes(rg *gin.RouterGroup, db *gorm.DB) {
    assetController := controllers.NewAssetController(db)
    notes := rg.Group("/notes")
    {
        notes.POST("", assetController.CreateNote)
        notes.GET("/:noteId", assetController.GetNote)
        notes.PUT("/:noteId", assetController.UpdateNote)
        notes.DELETE("/:noteId", assetController.DeleteNote)
        notes.POST("/:noteId/share", assetController.ShareNote)
    }
}