package application

import (
	"context"
	"errors"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"time"
)

// --- Folder Use Cases ---

type CreateFolderUseCase struct {
	folderRepo ports.FolderRepository
	eventBus   ports.EventPublisher
}

func NewCreateFolderUseCase(fr ports.FolderRepository, eb ports.EventPublisher) *CreateFolderUseCase {
	return &CreateFolderUseCase{folderRepo: fr, eventBus: eb}
}

type CreateFolderInput struct {
	Name    string
	OwnerID common.UserID
}

func (uc *CreateFolderUseCase) Execute(ctx context.Context, input CreateFolderInput) (*folder.Folder, error) {
	newFolder := &folder.Folder{
		Name:      input.Name,
		OwnerID:   input.OwnerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.folderRepo.Save(ctx, newFolder); err != nil {
		return nil, err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType: folder.FolderCreated,
		AssetType: "folder",
		AssetID:   newFolder.ID.String(),
		OwnerID:   newFolder.OwnerID.String(),
		ActionBy:  input.OwnerID.String(),
	})

	return newFolder, nil
}

type UpdateFolderUseCase struct {
	folderRepo ports.FolderRepository
	eventBus   ports.EventPublisher
}

func NewUpdateFolderUseCase(fr ports.FolderRepository, eb ports.EventPublisher) *UpdateFolderUseCase {
	return &UpdateFolderUseCase{folderRepo: fr, eventBus: eb}
}

type UpdateFolderInput struct {
	FolderID    common.FolderID
	Name        string
	RequesterID common.UserID
}

func (uc *UpdateFolderUseCase) Execute(ctx context.Context, input UpdateFolderInput) (*folder.Folder, error) {
	f, err := uc.folderRepo.FindByID(ctx, input.FolderID)
	if err != nil {
		return nil, err
	}
	if f.OwnerID != input.RequesterID {
		return nil, errors.New("user is not the owner of the folder")
	}

	f.Name = input.Name
	f.UpdatedAt = time.Now()

	if err := uc.folderRepo.Update(ctx, f); err != nil {
		return nil, err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType: folder.FolderUpdated,
		AssetType: "folder",
		AssetID:   f.ID.String(),
		OwnerID:   f.OwnerID.String(),
		ActionBy:  input.RequesterID.String(),
	})

	return f, nil
}

// --- Note Creation within a Folder ---

type CreateNoteInFolderUseCase struct {
	folderRepo ports.FolderRepository
	noteRepo   ports.NoteRepository
	eventBus   ports.EventPublisher
}

func NewCreateNoteInFolderUseCase(fr ports.FolderRepository, nr ports.NoteRepository, eb ports.EventPublisher) *CreateNoteInFolderUseCase {
	return &CreateNoteInFolderUseCase{folderRepo: fr, noteRepo: nr, eventBus: eb}
}

type CreateNoteInFolderInput struct {
	Title       string
	Body        string
	FolderID    common.FolderID
	RequesterID common.UserID
}

func (uc *CreateNoteInFolderUseCase) Execute(ctx context.Context, input CreateNoteInFolderInput) (*note.Note, error) {
	f, err := uc.folderRepo.FindByID(ctx, input.FolderID)
	if err != nil {
		return nil, err
	}
	if f.OwnerID != input.RequesterID {
		return nil, errors.New("user does not have write access to this folder")
	}

	newNote := &note.Note{
		Title:     input.Title,
		Body:      input.Body,
		FolderID:  input.FolderID,
		OwnerID:   input.RequesterID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.noteRepo.Save(ctx, newNote); err != nil {
		return nil, err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType: note.NoteCreated,
		AssetType: "note",
		AssetID:   newNote.ID.String(),
		OwnerID:   newNote.OwnerID.String(),
		ActionBy:  input.RequesterID.String(),
	})

	return newNote, nil
}

// --- Folder Sharing Use Cases ---

type ShareFolderUseCase struct {
	folderRepo ports.FolderRepository
	shareRepo  ports.ShareRepository
	eventBus   ports.EventPublisher
}

func NewShareFolderUseCase(fr ports.FolderRepository, sr ports.ShareRepository, eb ports.EventPublisher) *ShareFolderUseCase {
	return &ShareFolderUseCase{folderRepo: fr, shareRepo: sr, eventBus: eb}
}

type ShareFolderInput struct {
	FolderID    common.FolderID
	ShareWithID common.UserID
	AccessLevel common.Access
	RequesterID common.UserID
}

func (uc *ShareFolderUseCase) Execute(ctx context.Context, input ShareFolderInput) error {
	f, err := uc.folderRepo.FindByID(ctx, input.FolderID)
	if err != nil {
		return err
	}
	if f.OwnerID != input.RequesterID {
		return errors.New("only the owner can share a folder")
	}
	if !input.AccessLevel.IsValid() {
		return folder.ErrInvalidAccessLevel
	}

	share := &folder.Share{
		FolderID: input.FolderID,
		UserID:   input.ShareWithID,
		Access:   input.AccessLevel,
	}

	if err := uc.shareRepo.ShareFolder(ctx, share); err != nil {
		return err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType:    folder.FolderShared,
		AssetType:    "folder",
		AssetID:      input.FolderID.String(),
		OwnerID:      f.OwnerID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.ShareWithID.String(),
	})

	return nil
}

type UnshareFolderUseCase struct {
	folderRepo ports.FolderRepository
	shareRepo  ports.ShareRepository
	eventBus   ports.EventPublisher
}

func NewUnshareFolderUseCase(fr ports.FolderRepository, sr ports.ShareRepository, eb ports.EventPublisher) *UnshareFolderUseCase {
	return &UnshareFolderUseCase{folderRepo: fr, shareRepo: sr, eventBus: eb}
}

type UnshareFolderInput struct {
	FolderID      common.FolderID
	UnshareWithID common.UserID
	RequesterID   common.UserID
}

func (uc *UnshareFolderUseCase) Execute(ctx context.Context, input UnshareFolderInput) error {
	f, err := uc.folderRepo.FindByID(ctx, input.FolderID)
	if err != nil {
		return err
	}
	if f.OwnerID != input.RequesterID {
		return errors.New("only the owner can unshare a folder")
	}

	if err := uc.shareRepo.UnshareFolder(ctx, input.FolderID, input.UnshareWithID); err != nil {
		return err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType:    folder.FolderUnshared,
		AssetType:    "folder",
		AssetID:      input.FolderID.String(),
		OwnerID:      f.OwnerID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.UnshareWithID.String(),
	})

	return nil
}
