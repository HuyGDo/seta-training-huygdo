package routes

import (
	"seta/internal/app/server/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterFolderRoutes(rg *gin.RouterGroup, db *gorm.DB) {
    assetController := controllers.NewAssetController(db)
    folders := rg.Group("/folders")
    {
        folders.POST("", assetController.CreateFolder)
        folders.GET("/:folderId", assetController.GetFolder)
        folders.PUT("/:folderId", assetController.UpdateFolder)
        folders.DELETE("/:folderId", assetController.DeleteFolder)
        folders.POST("/:folderId/share", assetController.ShareFolder)
    }
}