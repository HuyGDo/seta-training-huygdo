package ports

import (
	"context"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"seta/internal/domain/team"
)

// TeamRepository defines the persistence operations for the Team aggregate.
type TeamRepository interface {
	Save(ctx context.Context, t *team.Team) error
	FindByID(ctx context.Context, id common.TeamID) (*team.Team, error)
	AddMember(ctx context.Context, teamID common.TeamID, member *team.Member) error
	RemoveMember(ctx context.Context, teamID common.TeamID, memberID common.UserID) error
	AddManager(ctx context.Context, teamID common.TeamID, manager *team.Manager) error
	RemoveManager(ctx context.Context, teamID common.TeamID, managerID common.UserID) error
	IsManager(ctx context.Context, teamID common.TeamID, userID common.UserID) (bool, error)
	IsLeadManager(ctx context.Context, teamID common.TeamID, userID common.UserID) (bool, error)
	GetTeamAssets(ctx context.Context, teamID common.TeamID) ([]folder.Folder, []note.Note, error)
}

// FolderRepository defines the persistence operations for the Folder aggregate.
type FolderRepository interface {
	Save(ctx context.Context, f *folder.Folder) error
	Update(ctx context.Context, f *folder.Folder) error
	Delete(ctx context.Context, id common.FolderID) error
	FindByID(ctx context.Context, id common.FolderID) (*folder.Folder, error)
}

// NoteRepository defines the persistence operations for the Note aggregate.
type NoteRepository interface {
	Save(ctx context.Context, n *note.Note) error
	Update(ctx context.Context, n *note.Note) error
	Delete(ctx context.Context, id common.NoteID) error
	FindByID(ctx context.Context, id common.NoteID) (*note.Note, error)
}

// ShareRepository defines persistence operations for sharing assets.
type ShareRepository interface {
	ShareFolder(ctx context.Context, share *folder.Share) error
	UnshareFolder(ctx context.Context, folderID common.FolderID, userID common.UserID) error
	ShareNote(ctx context.Context, share *note.Share) error
	UnshareNote(ctx context.Context, noteID common.NoteID, userID common.UserID) error
}

// UserRepository defines persistence operations related to users,
// specifically for querying data owned by them.
type UserRepository interface {
	FindUserAssets(ctx context.Context, userID common.UserID) ([]folder.Folder, []note.Note, error)
}
