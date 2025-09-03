package application

import (
	"context"
	"errors"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"seta/internal/domain/team"
	"time"
)

// --- Create Team Use Case ---

type CreateTeamUseCase struct {
	teamRepo  ports.TeamRepository
	eventBus  ports.EventPublisher
	txManager ports.TransactionManager
}

func NewCreateTeamUseCase(tr ports.TeamRepository, eb ports.EventPublisher, txm ports.TransactionManager) *CreateTeamUseCase {
	return &CreateTeamUseCase{teamRepo: tr, eventBus: eb, txManager: txm}
}

type CreateTeamInput struct {
	TeamName  string
	CreatorID common.UserID
	Managers  []team.Manager
	Members   []team.Member
}

func (uc *CreateTeamUseCase) Execute(ctx context.Context, input CreateTeamInput) (*team.Team, error) {
	var leadManagerCount int
	var isCreatorAManager bool
	for _, manager := range input.Managers {
		if manager.UserID == input.CreatorID {
			isCreatorAManager = true
		}
		if manager.IsLead {
			leadManagerCount++
		}
	}

	if !isCreatorAManager {
		return nil, team.ErrCreatorNotManager
	}
	if leadManagerCount != 1 {
		return nil, team.ErrTeamMustHaveLead
	}

	newTeam := &team.Team{
		Name:      input.TeamName,
		Managers:  input.Managers,
		Members:   input.Members,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := uc.txManager.Do(ctx, func(txCtx context.Context) error {
		return uc.teamRepo.Save(txCtx, newTeam)
	})
	if err != nil {
		return nil, err
	}

	go uc.eventBus.PublishTeamEvent(context.Background(), ports.EventPayload{
		EventType: team.TeamCreated,
		TeamID:    newTeam.ID.String(),
		ActionBy:  input.CreatorID.String(),
	})

	return newTeam, nil
}

// --- Add/Remove Member Use Cases ---

type AddMemberUseCase struct {
	teamRepo ports.TeamRepository
	eventBus ports.EventPublisher
}

func NewAddMemberUseCase(tr ports.TeamRepository, eb ports.EventPublisher) *AddMemberUseCase {
	return &AddMemberUseCase{teamRepo: tr, eventBus: eb}
}

type AddMemberInput struct {
	TeamID      common.TeamID
	MemberID    common.UserID
	RequesterID common.UserID
}

func (uc *AddMemberUseCase) Execute(ctx context.Context, input AddMemberInput) error {
	isManager, err := uc.teamRepo.IsManager(ctx, input.TeamID, input.RequesterID)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("requester is not a manager of this team")
	}

	member := &team.Member{UserID: input.MemberID}
	if err := uc.teamRepo.AddMember(ctx, input.TeamID, member); err != nil {
		return err
	}

	go uc.eventBus.PublishTeamEvent(context.Background(), ports.EventPayload{
		EventType:    team.MemberAdded,
		TeamID:       input.TeamID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.MemberID.String(),
	})
	return nil
}

type RemoveMemberUseCase struct {
	teamRepo ports.TeamRepository
	eventBus ports.EventPublisher
}

func NewRemoveMemberUseCase(tr ports.TeamRepository, eb ports.EventPublisher) *RemoveMemberUseCase {
	return &RemoveMemberUseCase{teamRepo: tr, eventBus: eb}
}

type RemoveMemberInput struct {
	TeamID      common.TeamID
	MemberID    common.UserID
	RequesterID common.UserID
}

func (uc *RemoveMemberUseCase) Execute(ctx context.Context, input RemoveMemberInput) error {
	isManager, err := uc.teamRepo.IsManager(ctx, input.TeamID, input.RequesterID)
	if err != nil {
		return err
	}
	if !isManager {
		return errors.New("requester is not a manager of this team")
	}

	if err := uc.teamRepo.RemoveMember(ctx, input.TeamID, input.MemberID); err != nil {
		return err
	}

	go uc.eventBus.PublishTeamEvent(context.Background(), ports.EventPayload{
		EventType:    team.MemberRemoved,
		TeamID:       input.TeamID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.MemberID.String(),
	})
	return nil
}

// --- Add/Remove Manager Use Cases ---

type AddManagerUseCase struct {
	teamRepo ports.TeamRepository
	eventBus ports.EventPublisher
}

func NewAddManagerUseCase(tr ports.TeamRepository, eb ports.EventPublisher) *AddManagerUseCase {
	return &AddManagerUseCase{teamRepo: tr, eventBus: eb}
}

type AddManagerInput struct {
	TeamID      common.TeamID
	ManagerID   common.UserID
	RequesterID common.UserID
}

func (uc *AddManagerUseCase) Execute(ctx context.Context, input AddManagerInput) error {
	isLead, err := uc.teamRepo.IsLeadManager(ctx, input.TeamID, input.RequesterID)
	if err != nil {
		return err
	}
	if !isLead {
		return errors.New("requester must be a lead manager to add other managers")
	}

	manager := &team.Manager{UserID: input.ManagerID, IsLead: false} // Can't add another lead
	if err := uc.teamRepo.AddManager(ctx, input.TeamID, manager); err != nil {
		return err
	}

	go uc.eventBus.PublishTeamEvent(context.Background(), ports.EventPayload{
		EventType:    team.ManagerAdded,
		TeamID:       input.TeamID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.ManagerID.String(),
	})
	return nil
}

type RemoveManagerUseCase struct {
	teamRepo ports.TeamRepository
	eventBus ports.EventPublisher
}

func NewRemoveManagerUseCase(tr ports.TeamRepository, eb ports.EventPublisher) *RemoveManagerUseCase {
	return &RemoveManagerUseCase{teamRepo: tr, eventBus: eb}
}

type RemoveManagerInput struct {
	TeamID      common.TeamID
	ManagerID   common.UserID
	RequesterID common.UserID
}

func (uc *RemoveManagerUseCase) Execute(ctx context.Context, input RemoveManagerInput) error {
	isLead, err := uc.teamRepo.IsLeadManager(ctx, input.TeamID, input.RequesterID)
	if err != nil {
		return err
	}
	if !isLead {
		return errors.New("requester must be a lead manager to remove other managers")
	}

	if err := uc.teamRepo.RemoveManager(ctx, input.TeamID, input.ManagerID); err != nil {
		return err
	}

	go uc.eventBus.PublishTeamEvent(context.Background(), ports.EventPayload{
		EventType:    team.ManagerRemoved,
		TeamID:       input.TeamID.String(),
		ActionBy:     input.RequesterID.String(),
		TargetUserID: input.ManagerID.String(),
	})
	return nil
}

// --- Get Team Assets Use Case ---

type GetTeamAssetsUseCase struct {
	teamRepo ports.TeamRepository
}

func NewGetTeamAssetsUseCase(tr ports.TeamRepository) *GetTeamAssetsUseCase {
	return &GetTeamAssetsUseCase{teamRepo: tr}
}

type GetTeamAssetsInput struct {
	TeamID      common.TeamID
	RequesterID common.UserID
}

type GetTeamAssetsOutput struct {
	Folders []folder.Folder
	Notes   []note.Note
}

func (uc *GetTeamAssetsUseCase) Execute(ctx context.Context, input GetTeamAssetsInput) (*GetTeamAssetsOutput, error) {
	isManager, err := uc.teamRepo.IsManager(ctx, input.TeamID, input.RequesterID)
	if err != nil {
		return nil, err
	}
	if !isManager {
		return nil, errors.New("requester is not a manager of this team")
	}

	// The repository will handle the caching logic internally
	folders, notes, err := uc.teamRepo.GetTeamAssets(ctx, input.TeamID)
	if err != nil {
		return nil, err
	}

	return &GetTeamAssetsOutput{Folders: folders, Notes: notes}, nil
}
