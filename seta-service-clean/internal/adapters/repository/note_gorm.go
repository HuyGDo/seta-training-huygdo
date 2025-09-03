package repository

import (
	"context"
	"errors"
	"seta/internal/domain/common"
	"seta/internal/domain/note"
	"seta/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormNoteRepository struct {
	db *gorm.DB
}

func NewGormNoteRepository(db *gorm.DB) *GormNoteRepository {
	return &GormNoteRepository{db: db}
}

func toDomainNote(gNote *models.GormNote) *note.Note {
	return &note.Note{
		ID:        common.NoteID(gNote.NoteID),
		Title:     gNote.Title,
		Body:      gNote.Body,
		FolderID:  common.FolderID(gNote.FolderID),
		OwnerID:   common.UserID(gNote.OwnerID),
		CreatedAt: gNote.CreatedAt,
		UpdatedAt: gNote.UpdatedAt,
	}
}

func fromDomainNote(dNote *note.Note) *models.GormNote {
	return &models.GormNote{
		NoteID:    uuid.UUID(dNote.ID),
		Title:     dNote.Title,
		Body:      dNote.Body,
		FolderID:  uuid.UUID(dNote.FolderID),
		OwnerID:   uuid.UUID(dNote.OwnerID),
		CreatedAt: dNote.CreatedAt,
		UpdatedAt: dNote.UpdatedAt,
	}
}

func (r *GormNoteRepository) Save(ctx context.Context, n *note.Note) error {
	gNote := fromDomainNote(n)
	gNote.NoteID = uuid.New()
	n.ID = common.NoteID(gNote.NoteID)

	return r.db.WithContext(ctx).Create(gNote).Error
}

func (r *GormNoteRepository) Update(ctx context.Context, n *note.Note) error {
	gNote := fromDomainNote(n)
	return r.db.WithContext(ctx).Model(&models.GormNote{}).Where("note_id = ?", gNote.NoteID).Updates(gNote).Error
}

func (r *GormNoteRepository) Delete(ctx context.Context, id common.NoteID) error {
	return r.db.WithContext(ctx).Delete(&models.GormNote{}, "note_id = ?", uuid.UUID(id)).Error
}

func (r *GormNoteRepository) FindByID(ctx context.Context, id common.NoteID) (*note.Note, error) {
	var gNote models.GormNote
	err := r.db.WithContext(ctx).First(&gNote, "note_id = ?", uuid.UUID(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, note.ErrNoteNotFound
		}
		return nil, err
	}
	return toDomainNote(&gNote), nil
}
