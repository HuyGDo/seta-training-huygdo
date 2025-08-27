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

// IsAssetOwner checks whether userID is the owner of the given asset.
func (s *AuthorizationService) IsAssetOwner(userID uuid.UUID, assetType string, assetID uuid.UUID) (bool, *errorHandling.CustomError) {
	type ownerRow struct {
		OwnerID uuid.UUID `gorm:"column:owner_id"`
	}

	var row ownerRow
	var err error

	switch assetType {
	case "folder":
		// NOTE: Don't use Pluck into scalar; use Select + Take into a typed struct
		err = s.db.Model(&models.Folder{}).
			Select("owner_id").
			Where("folder_id = ?", assetID).
			Take(&row).Error
	case "note":
		err = s.db.Model(&models.Note{}).
			Select("owner_id").
			Where("note_id = ?", assetID).
			Take(&row).Error
	default:
		return false, &errorHandling.CustomError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("invalid asset type: %s", assetType),
		}
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, &errorHandling.CustomError{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("%s not found", assetType),
			}
		}
		return false, &errorHandling.CustomError{
			Code:    http.StatusInternalServerError,
			Message: "Database error while checking ownership",
		}
	}

	return userID == row.OwnerID, nil
}

// CanAccessAsset checks whether userID has read or write access to the asset.
// Owner has implicit access.
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

// CanWriteAsset checks whether userID can write the asset.
// Owner has implicit write; otherwise requires share 'write'.
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

// getAccessFromCacheOrDB returns "", "read", or "write".
func (s *AuthorizationService) getAccessFromCacheOrDB(userID uuid.UUID, assetType string, assetID uuid.UUID) (string, *errorHandling.CustomError) {
	ctx := context.Background()
	aclCacheKey := "asset:" + assetID.String() + ":acl"

	// Try Redis cache first
	userAccess, err := database.Rdb.HGet(ctx, aclCacheKey, userID.String()).Result()
	if err == nil {
		return userAccess, nil
	}

	// Cache miss → fetch from DB and rebuild ACL
	acl, customErr := s.fetchAndBuildACL(assetType, assetID)
	if customErr != nil {
		return "", customErr
	}

	// Populate cache (if any entry)
	if len(acl) > 0 {
		// HSet map[string]interface{} is fine; set TTL for the hash key
		database.Rdb.HSet(ctx, aclCacheKey, acl)
		database.Rdb.Expire(ctx, aclCacheKey, 1*time.Hour)
	}

	// Return the specific user's access if present
	if access, ok := acl[userID.String()]; ok {
		if s, ok2 := access.(string); ok2 {
			return s, nil
		}
	}

	return "", nil
}

func (s *AuthorizationService) fetchAndBuildACL(assetType string, assetID uuid.UUID) (map[string]interface{}, *errorHandling.CustomError) {
	acl := make(map[string]interface{})

	switch assetType {
	case "folder":
		var shares []models.FolderShare
		if err := s.db.Where("folder_id = ?", assetID).Find(&shares).Error; err != nil {
			return nil, &errorHandling.CustomError{
				Code:    http.StatusInternalServerError,
				Message: "Database error checking folder shares",
			}
		}
		for _, share := range shares {
			acl[share.UserID.String()] = share.Access
		}
	case "note":
		var shares []models.NoteShare
		if err := s.db.Where("note_id = ?", assetID).Find(&shares).Error; err != nil {
			return nil, &errorHandling.CustomError{
				Code:    http.StatusInternalServerError,
				Message: "Database error checking note shares",
			}
		}
		for _, share := range shares {
			acl[share.UserID.String()] = share.Access
		}
	default:
		// Unknown asset type → return empty ACL (no access)
	}

	return acl, nil
}
