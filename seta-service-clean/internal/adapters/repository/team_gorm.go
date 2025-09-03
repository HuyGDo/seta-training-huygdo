package repository

import (
	"context"
	"errors"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"seta/internal/domain/team"
	"seta/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormTeamRepository struct {
	db *gorm.DB
}

func NewGormTeamRepository(db *gorm.DB) ports.TeamRepository {
	return &GormTeamRepository{db: db}
}

// toDomain maps a models.GormTeam to a domain Team entity.
func toDomain(gTeam *models.GormTeam) *team.Team {
	managers := make([]team.Manager, len(gTeam.Managers))
	for i, m := range gTeam.Managers {
		managers[i] = team.Manager{
			UserID: common.UserID(m.UserID),
			IsLead: m.IsLead,
		}
	}

	members := make([]team.Member, len(gTeam.Members))
	for i, m := range gTeam.Members {
		members[i] = team.Member{
			UserID: common.UserID(m.UserID),
		}
	}

	return &team.Team{
		ID:        common.TeamID(gTeam.ID),
		Name:      gTeam.TeamName,
		Managers:  managers,
		Members:   members,
		CreatedAt: gTeam.CreatedAt,
		UpdatedAt: gTeam.UpdatedAt,
	}
}

// fromDomain maps a domain Team entity to a models.GormTeam model.
func fromDomain(dTeam *team.Team) *models.GormTeam {
	gManagers := make([]models.GormTeamManager, len(dTeam.Managers))
	for i, m := range dTeam.Managers {
		gManagers[i] = models.GormTeamManager{UserID: uuid.UUID(m.UserID), IsLead: m.IsLead}
	}

	gMembers := make([]models.GormTeamMember, len(dTeam.Members))
	for i, m := range dTeam.Members {
		gMembers[i] = models.GormTeamMember{UserID: uuid.UUID(m.UserID)}
	}

	return &models.GormTeam{
		ID:        uuid.UUID(dTeam.ID),
		TeamName:  dTeam.Name,
		Managers:  gManagers,
		Members:   gMembers,
		CreatedAt: dTeam.CreatedAt,
		UpdatedAt: dTeam.UpdatedAt,
	}
}

func (r *GormTeamRepository) Save(ctx context.Context, t *team.Team) error {
	gTeam := fromDomain(t)
	if gTeam.ID == uuid.Nil {
		gTeam.ID = uuid.New()
		t.ID = common.TeamID(gTeam.ID) // Write back generated ID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(gTeam).Error; err != nil {
			return err
		}

		for i := range gTeam.Managers {
			gTeam.Managers[i].TeamID = gTeam.ID
		}
		if len(gTeam.Managers) > 0 {
			if err := tx.Create(&gTeam.Managers).Error; err != nil {
				return err
			}
		}

		for i := range gTeam.Members {
			gTeam.Members[i].TeamID = gTeam.ID
		}
		if len(gTeam.Members) > 0 {
			if err := tx.Create(&gTeam.Members).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormTeamRepository) FindByID(ctx context.Context, id common.TeamID) (*team.Team, error) {
	var gTeam models.GormTeam
	err := r.db.WithContext(ctx).Preload("Managers").Preload("Members").First(&gTeam, "id = ?", uuid.UUID(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}
	return toDomain(&gTeam), nil
}

func (r *GormTeamRepository) AddMember(ctx context.Context, teamID common.TeamID, member *team.Member) error {
	gMember := models.GormTeamMember{
		TeamID: uuid.UUID(teamID),
		UserID: uuid.UUID(member.UserID),
	}
	return r.db.WithContext(ctx).Create(&gMember).Error
}

func (r *GormTeamRepository) RemoveMember(ctx context.Context, teamID common.TeamID, memberID common.UserID) error {
	return r.db.WithContext(ctx).Delete(&models.GormTeamMember{}, "team_id = ? AND user_id = ?", uuid.UUID(teamID), uuid.UUID(memberID)).Error
}

func (r *GormTeamRepository) AddManager(ctx context.Context, teamID common.TeamID, manager *team.Manager) error {
	gManager := models.GormTeamManager{
		TeamID: uuid.UUID(teamID),
		UserID: uuid.UUID(manager.UserID),
		IsLead: manager.IsLead,
	}
	return r.db.WithContext(ctx).Create(&gManager).Error
}

func (r *GormTeamRepository) RemoveManager(ctx context.Context, teamID common.TeamID, managerID common.UserID) error {
	return r.db.WithContext(ctx).Delete(&models.GormTeamManager{}, "team_id = ? AND user_id = ?", uuid.UUID(teamID), uuid.UUID(managerID)).Error
}

func (r *GormTeamRepository) IsManager(ctx context.Context, teamID common.TeamID, userID common.UserID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.GormTeamManager{}).Where("team_id = ? AND user_id = ?", uuid.UUID(teamID), uuid.UUID(userID)).Count(&count).Error
	return count > 0, err
}

func (r *GormTeamRepository) IsLeadManager(ctx context.Context, teamID common.TeamID, userID common.UserID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.GormTeamManager{}).Where("team_id = ? AND user_id = ? AND is_lead = ?", uuid.UUID(teamID), uuid.UUID(userID), true).Count(&count).Error
	return count > 0, err
}

func (r *GormTeamRepository) GetTeamAssets(ctx context.Context, teamID common.TeamID) ([]folder.Folder, []note.Note, error) {
	var memberIDs []uuid.UUID
	if err := r.db.WithContext(ctx).Model(&models.GormTeamMember{}).Where("team_id = ?", uuid.UUID(teamID)).Pluck("user_id", &memberIDs).Error; err != nil {
		return nil, nil, err
	}

	if len(memberIDs) == 0 {
		return []folder.Folder{}, []note.Note{}, nil
	}

	var gFolders []models.GormFolder
	if err := r.db.WithContext(ctx).
		Joins("LEFT JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folders.owner_id IN (?) OR folder_shares.user_id IN (?)", memberIDs, memberIDs).
		Group("folders.folder_id").
		Find(&gFolders).Error; err != nil {
		return nil, nil, err
	}

	var gNotes []models.GormNote
	if err := r.db.WithContext(ctx).
		Joins("LEFT JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("notes.owner_id IN (?) OR note_shares.user_id IN (?)", memberIDs, memberIDs).
		Group("notes.note_id").
		Find(&gNotes).Error; err != nil {
		return nil, nil, err
	}

	domainFolders := make([]folder.Folder, len(gFolders))
	for i, f := range gFolders {
		domainFolders[i] = *models.ToDomainFolder(&f)
	}

	domainNotes := make([]note.Note, len(gNotes))
	for i, n := range gNotes {
		domainNotes[i] = *models.ToDomainNote(&n)
	}

	return domainFolders, domainNotes, nil
}
