package models

import (
	"time"

	"github.com/google/uuid"
)

// Folder represents a folder in the system.
type Folder struct {
	FolderID  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"folderId"`
	Name      string    `gorm:"not null" json:"name"`
	OwnerID   uuid.UUID `gorm:"type:uuid" json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Folder) TableName() string {
	return "folders"
}

// Note represents a note in the system.
type Note struct {
	NoteID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"noteId"`
	Title     string    `gorm:"not null" json:"title"`
	Body      string    `json:"body"`
	FolderID  uuid.UUID `gorm:"type:uuid" json:"folderId"`
	OwnerID   uuid.UUID `gorm:"type:uuid" json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Note) TableName() string {
	return "notes"
}

// FolderShare represents the sharing of a folder with a user.
type FolderShare struct {
	FolderID uuid.UUID `gorm:"type:uuid;primaryKey" json:"folderId"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Access   string    `gorm:"not null" json:"access"` // "read" or "write"
}

func (FolderShare) TableName() string {
	return "folder_shares"
}

// NoteShare represents the sharing of a note with a user.
type NoteShare struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey" json:"noteId"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Access string    `gorm:"not null" json:"access"` // "read" or "write"
}

func (NoteShare) TableName() string {
	return "note_shares"
}