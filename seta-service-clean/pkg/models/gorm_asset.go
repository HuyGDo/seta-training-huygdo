package models

import (
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"time"

	"github.com/google/uuid"
)

// gormFolder is the GORM model for a folder.
type GormFolder struct {
	FolderID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"not null"`
	OwnerID   uuid.UUID `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (GormFolder) TableName() string { return "folders" }

// ToDomainFolder maps a gormFolder to a domain Folder entity.
func ToDomainFolder(gFolder *GormFolder) *folder.Folder {
	return &folder.Folder{
		ID:        common.FolderID(gFolder.FolderID),
		Name:      gFolder.Name,
		OwnerID:   common.UserID(gFolder.OwnerID),
		CreatedAt: gFolder.CreatedAt,
		UpdatedAt: gFolder.UpdatedAt,
	}
}

// GormNote is the GORM model for a note.
type GormNote struct {
	NoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title     string    `gorm:"not null"`
	Body      string
	FolderID  uuid.UUID `gorm:"type:uuid"`
	OwnerID   uuid.UUID `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (GormNote) TableName() string { return "notes" }

// ToDomainNote maps a GormNote to a domain Note entity.
func ToDomainNote(gNote *GormNote) *note.Note {
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

// gormFolderShare is the GORM model for a folder share.
type GormFolderShare struct {
	FolderID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Access   string    `gorm:"not null"`
}

func (GormFolderShare) TableName() string { return "folder_shares" }

// gormNoteShare is the GORM model for a note share.
type GormNoteShare struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Access string    `gorm:"not null"`
}

func (GormNoteShare) TableName() string { return "note_shares" }
