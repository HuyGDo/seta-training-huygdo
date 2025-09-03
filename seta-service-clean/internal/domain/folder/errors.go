package folder

import "errors"

var (
	// ErrFolderNotFound is returned when a folder cannot be found.
	ErrFolderNotFound = errors.New("folder not found")
	// ErrInvalidAccessLevel is returned when an invalid access level is provided for a share.
	ErrInvalidAccessLevel = errors.New("invalid access level provided")
)
