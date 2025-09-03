package note

import "errors"

var (
	// ErrNoteNotFound is returned when a note cannot be found.
	ErrNoteNotFound = errors.New("note not found")
)