package note

import (
	"seta/internal/domain/common"
	"time"
)

// Note is an aggregate root representing a single note within a folder.
type Note struct {
	ID        common.NoteID
	Title     string
	Body      string
	FolderID  common.FolderID
	OwnerID   common.UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Share represents the permissions a user has on a specific note.
type Share struct {
	NoteID common.NoteID
	UserID common.UserID
	Access common.Access
}
