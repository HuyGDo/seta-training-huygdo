package folder

import (
	"seta/internal/domain/common"
	"time"
)

// Folder is an aggregate root representing a container for notes.
type Folder struct {
	ID        common.FolderID
	Name      string
	OwnerID   common.UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Share represents the permissions a user has on a specific folder.
type Share struct {
	FolderID common.FolderID
	UserID   common.UserID
	Access   common.Access
}
