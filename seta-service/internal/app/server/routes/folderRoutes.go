package routes

import (
	"seta/internal/app/server/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterFolderRoutes(rg *gin.RouterGroup, db *gorm.DB) {
    folderController := controllers.NewFolderController(db)
    folders := rg.Group("/folders")
    {
        folders.POST("", folderController.CreateFolder)
        folders.GET("/:folderId", folderController.GetFolder)
        folders.PUT("/:folderId", folderController.UpdateFolder)
        folders.DELETE("/:folderId", folderController.DeleteFolder)
        folders.POST("/:folderId/share", folderController.ShareFolder)
		folders.DELETE("/:folderId/share/:userId", folderController.RevokeFolderSharing)
    }
}