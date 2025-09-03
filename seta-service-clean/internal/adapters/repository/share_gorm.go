package repository

import (
	"context"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"seta/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormShareRepository struct {
	db *gorm.DB
}

func NewGormShareRepository(db *gorm.DB) *GormShareRepository {
	return &GormShareRepository{db: db}
}

func (r *GormShareRepository) ShareFolder(ctx context.Context, share *folder.Share) error {
	gShare := models.GormFolderShare{
		FolderID: uuid.UUID(share.FolderID),
		UserID:   uuid.UUID(share.UserID),
		Access:   string(share.Access),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&gShare).Error
}

func (r *GormShareRepository) UnshareFolder(ctx context.Context, folderID common.FolderID, userID common.UserID) error {
	return r.db.WithContext(ctx).Delete(&models.GormFolderShare{}, "folder_id = ? AND user_id = ?", uuid.UUID(folderID), uuid.UUID(userID)).Error
}

func (r *GormShareRepository) ShareNote(ctx context.Context, share *note.Share) error {
	gShare := models.GormNoteShare{
		NoteID: uuid.UUID(share.NoteID),
		UserID: uuid.UUID(share.UserID),
		Access: string(share.Access),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&gShare).Error
}

func (r *GormShareRepository) UnshareNote(ctx context.Context, noteID common.NoteID, userID common.UserID) error {
	return r.db.WithContext(ctx).Delete(&models.GormNoteShare{}, "note_id = ? AND user_id = ?", uuid.UUID(noteID), uuid.UUID(userID)).Error
}
