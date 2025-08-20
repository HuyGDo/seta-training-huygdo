package services

import (
	"errors"
	"fmt"
	"net/http"
	"seta/internal/pkg/errorHandling"
	"seta/internal/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthorizationService struct {
	db *gorm.DB
}

func NewAuthorizationService(db *gorm.DB) *AuthorizationService {
	return &AuthorizationService{db: db}
}

// IsAssetOwner is now updated to return *errorHandling.CustomError consistently.
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

// CanAccessAsset is updated to correctly handle the custom error from IsAssetOwner.
func (s *AuthorizationService) CanAccessAsset(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	isOwner, err := s.IsAssetOwner(userID, assetType, assetID)
	if err != nil || isOwner {
		return isOwner, err
	}

	switch assetType {
	case "folder":
		var count int64
		if dbErr := s.db.Model(&models.FolderShare{}).Where("folder_id = ? AND user_id = ?", assetID, userID).Count(&count).Error; dbErr != nil {
			return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking folder share"}
		}
		return count > 0, nil

	case "note":
		var count int64
		if dbErr := s.db.Model(&models.NoteShare{}).Where("note_id = ? AND user_id = ?", assetID, userID).Count(&count).Error; dbErr != nil {
			return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking note share"}
		}
		if count > 0 {
			return true, nil
		}

		var note models.Note
		s.db.Select("folder_id").First(&note, "note_id = ?", assetID)
		return s.CanAccessAsset(userID, "folder", note.FolderID)
	}

	return false, nil
}

// CanWriteAsset is also updated to correctly handle the custom error.
func (s *AuthorizationService) CanWriteAsset(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	isOwner, err := s.IsAssetOwner(userID, assetType, assetID)
	if err != nil || isOwner {
		return isOwner, err
	}

	switch assetType {
	case "folder":
		var count int64
		if dbErr := s.db.Model(&models.FolderShare{}).Where("folder_id = ? AND user_id = ? AND access = 'write'", assetID, userID).Count(&count).Error; dbErr != nil {
			return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking folder write access"}
		}
		return count > 0, nil

	case "note":
		var count int64
		if dbErr := s.db.Model(&models.NoteShare{}).Where("note_id = ? AND user_id = ? AND access = 'write'", assetID, userID).Count(&count).Error; dbErr != nil {
			return false, &errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Database error checking note write access"}
		}
		if count > 0 {
			return true, nil
		}

		var note models.Note
		s.db.Select("folder_id").First(&note, "note_id = ?", assetID)
		return s.CanWriteAsset(userID, "folder", note.FolderID)
	}

	return false, nil
}