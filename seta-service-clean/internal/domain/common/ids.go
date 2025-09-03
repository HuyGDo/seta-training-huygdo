package common

import "github.com/google/uuid"

// UserID is a strongly-typed identifier for a user.
type UserID uuid.UUID

// TeamID is a strongly-typed identifier for a team.
type TeamID uuid.UUID

// FolderID is a strongly-typed identifier for a folder.
type FolderID uuid.UUID

// NoteID is a strongly-typed identifier for a note.
type NoteID uuid.UUID

// String returns the string representation of the UserID.
func (id UserID) String() string {
	return uuid.UUID(id).String()
}

// String returns the string representation of the TeamID.
func (id TeamID) String() string {
	return uuid.UUID(id).String()
}

// String returns the string representation of the FolderID.
func (id FolderID) String() string {
	return uuid.UUID(id).String()
}

// String returns the string representation of the NoteID.
func (id NoteID) String() string {
	return uuid.UUID(id).String()
}
