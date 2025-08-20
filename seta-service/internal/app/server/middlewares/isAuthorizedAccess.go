package middlewares

import (
	"net/http"
	"seta/internal/app/server/services"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAccessMiddleware now uses the centralized utility functions for all ID parsing.
func AssetAccessMiddleware(assetType string, assetIDParamName string, checkFunc func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError), db *gorm.DB) gin.HandlerFunc {
	authorization := services.NewAuthorizationService(db)

	return func(c *gin.Context) {
		
		assetID, err := utils.GetUUIDFromParam(c, assetIDParamName)
		if err != nil {
			_ = c.Error(err) // Pass the structured error to the error handler
			c.Abort()
			return
		}

		userID, err := utils.GetUserUUIDFromContext(c)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		// The core permission logic remains the same.
		// hasPermission, err := checkFunc(authorization, userID, assetID)
		// if err.Code != 0 {
		// 	_ = c.Error(err)
		// 	c.Abort()
		// 	return
		// }

		hasPermission, customErr := checkFunc(authorization, userID, assetID)
		if customErr.Code != 0 {
			_ = c.Error(customErr)
			c.Abort()
			return
		}
			

		if !hasPermission {
			// The error is now handled by the centralized error middleware
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusForbidden, Message: "You are not authorized for this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}


func CanReadNote(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("note", "noteId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.CanAccessAsset(userID, "note", assetID)
		}, db)
}

func CanWriteNote(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("note", "noteId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.CanWriteAsset(userID, "note", assetID)
		}, db)
}

func IsNoteOwner(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("note", "noteId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.IsAssetOwner(userID, "note", assetID)
		}, db)
}

func CanReadFolder(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("folder", "folderId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.CanAccessAsset(userID, "folder", assetID)
		}, db)
}

func CanWriteFolder(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("folder", "folderId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.CanWriteAsset(userID, "folder", assetID)
		}, db)
}

func IsFolderOwner(db *gorm.DB) gin.HandlerFunc {
	return AssetAccessMiddleware("folder", "folderId",
		func(authorization *services.AuthorizationService, userID, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
			return authorization.IsAssetOwner(userID, "folder", assetID)
		}, db)
}