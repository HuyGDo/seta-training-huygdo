package repository

import (
	"context"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"seta/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserRepository implements the UserRepository port.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new instance of GormUserRepository.
func NewGormUserRepository(db *gorm.DB) ports.UserRepository {
	return &GormUserRepository{db: db}
}

// FindUserAssets retrieves all folders and notes that a user either owns or has been shared with.
func (r *GormUserRepository) FindUserAssets(ctx context.Context, userID common.UserID) ([]folder.Folder, []note.Note, error) {
	var gFolders []models.GormFolder
	uid := uuid.UUID(userID)

	// Find all folders the user owns OR has a direct share on.
	if err := r.db.WithContext(ctx).
		Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id = ? OR folder_shares.user_id = ?", uid, uid).
		Group("folders.folder_id").
		Find(&gFolders).Error; err != nil {
		return nil, nil, err
	}

	var gNotes []models.GormNote
	// Find all notes the user owns, has a direct share on, OR are in a folder shared with them.
	if err := r.db.WithContext(ctx).
		Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Joins("LEFT JOIN folder_shares ON notes.folder_id = folder_shares.folder_id").
		Where("notes.owner_id = ? OR note_shares.user_id = ? OR folder_shares.user_id = ?", uid, uid, uid).
		Group("notes.note_id").
		Find(&gNotes).Error; err != nil {
		return nil, nil, err
	}

	// Map GORM models to domain entities
	domainFolders := make([]folder.Folder, len(gFolders))
	for i, f := range gFolders {
		domainFolders[i] = *models.ToDomainFolder(&f)
	}

	domainNotes := make([]note.Note, len(gNotes))
	for i, n := range gNotes {
		domainNotes[i] = *models.ToDomainNote(&n)
	}

	return domainFolders, domainNotes, nil
}
