package application

import (
	"context"
	"errors"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/note"
	"time"
)

// --- Note Use Cases ---

type GetNoteUseCase struct {
	noteRepo ports.NoteRepository
}

func NewGetNoteUseCase(nr ports.NoteRepository) *GetNoteUseCase {
	return &GetNoteUseCase{noteRepo: nr}
}

type GetNoteInput struct {
	NoteID      common.NoteID
	RequesterID common.UserID
}

func (uc *GetNoteUseCase) Execute(ctx context.Context, input GetNoteInput) (*note.Note, error) {
	n, err := uc.noteRepo.FindByID(ctx, input.NoteID)
	if err != nil {
		return nil, err
	}

	if n.OwnerID != input.RequesterID {
		return nil, errors.New("user does not have permission to read this note")
	}

	return n, nil
}

type UpdateNoteUseCase struct {
	noteRepo ports.NoteRepository
	eventBus ports.EventPublisher
}

func NewUpdateNoteUseCase(nr ports.NoteRepository, eb ports.EventPublisher) *UpdateNoteUseCase {
	return &UpdateNoteUseCase{noteRepo: nr, eventBus: eb}
}

type UpdateNoteInput struct {
	NoteID      common.NoteID
	Title       string
	Body        string
	RequesterID common.UserID
}

func (uc *UpdateNoteUseCase) Execute(ctx context.Context, input UpdateNoteInput) (*note.Note, error) {
	n, err := uc.noteRepo.FindByID(ctx, input.NoteID)
	if err != nil {
		return nil, err
	}

	if n.OwnerID != input.RequesterID {
		return nil, errors.New("user does not have permission to write to this note")
	}

	n.Title = input.Title
	n.Body = input.Body
	n.UpdatedAt = time.Now()

	if err := uc.noteRepo.Update(ctx, n); err != nil {
		return nil, err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType: note.NoteUpdated,
		AssetType: "note",
		AssetID:   n.ID.String(),
		OwnerID:   n.OwnerID.String(),
		ActionBy:  input.RequesterID.String(),
	})

	return n, nil
}

type DeleteNoteUseCase struct {
	noteRepo ports.NoteRepository
	eventBus ports.EventPublisher
}

func NewDeleteNoteUseCase(nr ports.NoteRepository, eb ports.EventPublisher) *DeleteNoteUseCase {
	return &DeleteNoteUseCase{noteRepo: nr, eventBus: eb}
}

type DeleteNoteInput struct {
	NoteID      common.NoteID
	RequesterID common.UserID
}

func (uc *DeleteNoteUseCase) Execute(ctx context.Context, input DeleteNoteInput) error {
	n, err := uc.noteRepo.FindByID(ctx, input.NoteID)
	if err != nil {
		return err
	}

	if n.OwnerID != input.RequesterID {
		return errors.New("only the owner can delete a note")
	}

	if err := uc.noteRepo.Delete(ctx, input.NoteID); err != nil {
		return err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType: note.NoteDeleted,
		AssetType: "note",
		AssetID:   n.ID.String(),
		OwnerID:   n.OwnerID.String(),
		ActionBy:  input.RequesterID.String(),
	})

	return nil
}

// --- Note Sharing Use Cases ---

type ShareNoteUseCase struct {
	noteRepo  ports.NoteRepository
	shareRepo ports.ShareRepository
	eventBus  ports.EventPublisher
}

func NewShareNoteUseCase(nr ports.NoteRepository, sr ports.ShareRepository, eb ports.EventPublisher) *ShareNoteUseCase {
	return &ShareNoteUseCase{noteRepo: nr, shareRepo: sr, eventBus: eb}
}

type ShareNoteInput struct {
	NoteID      common.NoteID
	ShareWithID common.UserID
	AccessLevel common.Access
	RequesterID common.UserID
}

func (uc *ShareNoteUseCase) Execute(ctx context.Context, input ShareNoteInput) error {
	n, err := uc.noteRepo.FindByID(ctx, input.NoteID)
	if err != nil {
		return err
	}
	if n.OwnerID != input.RequesterID {
		return errors.New("only the owner can share a note")
	}
	if !input.AccessLevel.IsValid() {
		return errors.New("invalid access level provided")
	}

	share := &note.Share{
		NoteID: input.NoteID,
		UserID:   input.ShareWithID,
		Access:   input.AccessLevel,
	}

	if err := uc.shareRepo.ShareNote(ctx, share); err != nil {
		return err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType:    note.NoteShared,
		AssetType:    "note",
		AssetID:      input.NoteID.String(),
		OwnerID:      n.OwnerID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.ShareWithID.String(),
	})

	return nil
}

type UnshareNoteUseCase struct {
	noteRepo  ports.NoteRepository
	shareRepo ports.ShareRepository
	eventBus  ports.EventPublisher
}

func NewUnshareNoteUseCase(nr ports.NoteRepository, sr ports.ShareRepository, eb ports.EventPublisher) *UnshareNoteUseCase {
	return &UnshareNoteUseCase{noteRepo: nr, shareRepo: sr, eventBus: eb}
}

type UnshareNoteInput struct {
	NoteID        common.NoteID
	UnshareWithID common.UserID
	RequesterID   common.UserID
}

func (uc *UnshareNoteUseCase) Execute(ctx context.Context, input UnshareNoteInput) error {
	n, err := uc.noteRepo.FindByID(ctx, input.NoteID)
	if err != nil {
		return err
	}
	if n.OwnerID != input.RequesterID {
		return errors.New("only the owner can unshare a note")
	}

	if err := uc.shareRepo.UnshareNote(ctx, input.NoteID, input.UnshareWithID); err != nil {
		return err
	}

	go uc.eventBus.PublishAssetEvent(context.Background(), ports.EventPayload{
		EventType:    note.NoteUnshared,
		AssetType:    "note",
		AssetID:      input.NoteID.String(),
		OwnerID:      n.OwnerID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.UnshareWithID.String(),
	})

	return nil
}
