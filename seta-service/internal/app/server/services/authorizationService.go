package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"seta/internal/pkg/database"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthorizationService struct {
	db *gorm.DB
}

func NewAuthorizationService(db *gorm.DB) *AuthorizationService {
	return &AuthorizationService{db: db}
}

func (s *AuthorizationService) IsAssetOwner(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	var ownerID uuid.UUID
	var err error

	switch assetType {
	case "folder":
		err = s.db.Model(&models.Folder{}).Where("folder_id = ?", assetID).Pluck("owner_id", &ownerID).Error
	case "note":
		err = s.db.Model(&models.Note{}).Where("note_id = ?", assetID).Pluck("owner_id", &ownerID).Error
	default:
		return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("invalid asset type: %s", assetType)}
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, &errorHandling.CustomError{Code: http.StatusNotFound, Message: fmt.Sprintf("%s not found", assetType)}
		}
		return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error while checking ownership"}
	}

	return userID == ownerID, nil
}

func (s *AuthorizationService) CanAccessAsset(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	isOwner, ownerErr := s.IsAssetOwner(userID, assetType, assetID)
	if ownerErr != nil || isOwner {
		return isOwner, ownerErr
	}

	access, err := s.getAccessFromCacheOrDB(userID, assetType, assetID)
	if err != nil {
		return false, err
	}

	return access == "read" || access == "write", nil
}

func (s *AuthorizationService) CanWriteAsset(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	isOwner, ownerErr := s.IsAssetOwner(userID, assetType, assetID)
	if ownerErr != nil || isOwner {
		return isOwner, ownerErr
	}

	access, err := s.getAccessFromCacheOrDB(userID, assetType, assetID)
	if err != nil {
		return false, err
	}

	return access == "write", nil
}

func (s *AuthorizationService) getAccessFromCacheOrDB(userID uuid.UUID, assetType string, assetID uuid.UUID) (string, *errorHandling.CustomError) {
	ctx := context.Background()
	aclCacheKey := "asset:" + assetID.String() + ":acl"

	userAccess, err := database.Rdb.HGet(ctx, aclCacheKey, userID.String()).Result()
	if err == nil {
		return userAccess, nil
	}

	acl, customErr := s.fetchAndBuildACL(assetType, assetID)
	if customErr != nil {
		return "", customErr
	}

	if len(acl) > 0 {
		database.Rdb.HSet(ctx, aclCacheKey, acl)
		database.Rdb.Expire(ctx, aclCacheKey, 1*time.Hour)
	}

	if access, ok := acl[userID.String()]; ok {
		return access.(string), nil
	}

	return "", nil
}

func (s *AuthorizationService) fetchAndBuildACL(assetType string, assetID uuid.UUID) (map[string]interface{}, *errorHandling.CustomError) {
	acl := make(map[string]interface{})

	switch assetType {
	case "folder":
		var shares []models.FolderShare
		if err := s.db.Where("folder_id = ?", assetID).Find(&shares).Error; err != nil {
			return nil, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking folder shares"}
		}
		for _, share := range shares {
			acl[share.UserID.String()] = share.Access
		}
	case "note":
		var shares []models.NoteShare
		if err := s.db.Where("note_id = ?", assetID).Find(&shares).Error; err != nil {
			return nil, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking note shares"}
		}
		for _, share := range shares {
			acl[share.UserID.String()] = share.Access
		}
	}

	return acl, nil
}