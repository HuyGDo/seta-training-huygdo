package routes

import (
	"seta/internal/app/server/controllers"
	"seta/internal/app/server/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterFolderRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	folderController := controllers.NewFolderController(db)
	folders := rg.Group("/folders")
	{
		// No asset auth needed, just auth from the parent router group.
		folders.POST("", folderController.CreateFolder)

		// Routes requiring specific permissions on an existing folder.
		folders.GET("/:folderId", middlewares.CanReadFolder(db), folderController.GetFolder)
		folders.PUT("/:folderId", middlewares.CanWriteFolder(db), folderController.UpdateFolder)
		folders.DELETE("/:folderId", middlewares.IsFolderOwner(db), folderController.DeleteFolder)
		folders.POST("/:folderId/share", middlewares.IsFolderOwner(db), folderController.ShareFolder)
		folders.DELETE("/:folderId/share/:userId", middlewares.IsFolderOwner(db), folderController.RevokeFolderSharing)

		// To create a note in a folder, the user needs write access to it.
		folders.POST("/:folderId/notes", middlewares.CanWriteFolder(db), folderController.CreateNote)
	}
}