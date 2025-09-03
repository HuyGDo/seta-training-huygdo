package ports

import (
	"context"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"

	"github.com/google/uuid"
)

// TeamMemberCache defines the interface for caching team member IDs.
type TeamMemberCache interface {
	GetTeamMembers(ctx context.Context, teamID string) ([]common.UserID, error)
	SetTeamMembers(ctx context.Context, teamID string, memberIDs []common.UserID) error
	AddTeamMember(ctx context.Context, teamID string, memberID string) error
	RemoveTeamMember(ctx context.Context, teamID string, memberID string) error
}

// AssetMetaCache defines the interface for caching asset metadata (folders and notes).
type AssetMetaCache interface {
	GetFolder(ctx context.Context, folderID common.FolderID) (*folder.Folder, error)
	SetFolder(ctx context.Context, f *folder.Folder) error
	InvalidateFolder(ctx context.Context, folderID common.FolderID) error
	GetNote(ctx context.Context, noteID common.NoteID) (*note.Note, error)
	SetNote(ctx context.Context, n *note.Note) error
	InvalidateNote(ctx context.Context, noteID common.NoteID) error
}

// ACLCache defines the interface for caching Access Control Lists for assets.
type ACLCache interface {
	GetACL(ctx context.Context, assetID uuid.UUID) (map[common.UserID]common.Access, error)
	SetACL(ctx context.Context, assetID uuid.UUID, acl map[common.UserID]common.Access) error
	InvalidateACL(ctx context.Context, assetID uuid.UUID) error
}
