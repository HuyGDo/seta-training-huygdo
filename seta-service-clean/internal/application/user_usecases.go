package application

import (
	"context"
	"errors"
	"io"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
)

// ImportUsersUseCase orchestrates the CSV import process by delegating to an importer.
type ImportUsersUseCase struct {
	userImporter ports.UserImporter
}

// NewImportUsersUseCase creates a new instance.
func NewImportUsersUseCase(importer ports.UserImporter) *ImportUsersUseCase {
	return &ImportUsersUseCase{userImporter: importer}
}

// Execute handles the business logic of importing users from a CSV file.
func (uc *ImportUsersUseCase) Execute(ctx context.Context, file io.Reader) (ports.UserImportSummary, error) {
	if file == nil {
		return ports.UserImportSummary{}, errors.New("input file cannot be nil")
	}
	// The use case's responsibility is simply to orchestrate the import.
	// All the complex logic is handled by the adapter that implements the UserImporter port.
	return uc.userImporter.ImportUsers(ctx, file)
}

// ---

// GetUserAssetsUseCase is responsible for fetching all assets a user has access to.
type GetUserAssetsUseCase struct {
	userRepo ports.UserRepository
}

// NewGetUserAssetsUseCase creates a new instance.
func NewGetUserAssetsUseCase(ur ports.UserRepository) *GetUserAssetsUseCase {
	return &GetUserAssetsUseCase{userRepo: ur}
}

// GetUserAssetsInput is the DTO for this use case.
type GetUserAssetsInput struct {
	TargetUserID  common.UserID // The user whose assets we want to see
	RequesterID   common.UserID // The user making the request
	RequesterRole string
}

type GetUserAssetsOutput struct {
	Folders []folder.Folder
	Notes   []note.Note
}

// Execute runs the use case.
func (uc *GetUserAssetsUseCase) Execute(ctx context.Context, input GetUserAssetsInput) (*GetUserAssetsOutput, error) {
	// 1. Authorization Business Rule
	if input.TargetUserID != input.RequesterID {
		return nil, errors.New("you are not authorized to view these assets")
	}

	// 2. Delegate data fetching to the repository
	folders, notes, err := uc.userRepo.FindUserAssets(ctx, input.TargetUserID)
	if err != nil {
		return nil, err // Propagate persistence errors
	}

	// 3. Return the result
	output := &GetUserAssetsOutput{
		Folders: folders,
		Notes:   notes,
	}

	return output, nil
}
