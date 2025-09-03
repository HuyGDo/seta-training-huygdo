package repository

import (
	"context"
	"errors"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormFolderRepository struct {
	db *gorm.DB
}

func NewGormFolderRepository(db *gorm.DB) *GormFolderRepository {
	return &GormFolderRepository{db: db}
}

func toDomainFolder(gFolder *models.GormFolder) *folder.Folder {
	return &folder.Folder{
		ID:        common.FolderID(gFolder.FolderID),
		Name:      gFolder.Name,
		OwnerID:   common.UserID(gFolder.OwnerID),
		CreatedAt: gFolder.CreatedAt,
		UpdatedAt: gFolder.UpdatedAt,
	}
}

func fromDomainFolder(dFolder *folder.Folder) *models.GormFolder {
	return &models.GormFolder{
		FolderID:  uuid.UUID(dFolder.ID),
		Name:      dFolder.Name,
		OwnerID:   uuid.UUID(dFolder.OwnerID),
		CreatedAt: dFolder.CreatedAt,
		UpdatedAt: dFolder.UpdatedAt,
	}
}

func (r *GormFolderRepository) Save(ctx context.Context, f *folder.Folder) error {
	gFolder := fromDomainFolder(f)
	gFolder.FolderID = uuid.New() // Generate ID
	f.ID = common.FolderID(gFolder.FolderID) // Write back the generated ID

	return r.db.WithContext(ctx).Create(gFolder).Error
}

func (r *GormFolderRepository) Update(ctx context.Context, f *folder.Folder) error {
	gFolder := fromDomainFolder(f)
	return r.db.WithContext(ctx).Model(&models.GormFolder{}).Where("folder_id = ?", gFolder.FolderID).Updates(gFolder).Error
}

func (r *GormFolderRepository) Delete(ctx context.Context, id common.FolderID) error {
	// Deleting shares and notes within the folder should be handled by the use case transaction
	return r.db.WithContext(ctx).Delete(&models.GormFolder{}, "folder_id = ?", uuid.UUID(id)).Error
}

func (r *GormFolderRepository) FindByID(ctx context.Context, id common.FolderID) (*folder.Folder, error) {
	var gFolder models.GormFolder
	err := r.db.WithContext(ctx).First(&gFolder, "folder_id = ?", uuid.UUID(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, folder.ErrFolderNotFound
		}
		return nil, err
	}
	return toDomainFolder(&gFolder), nil
}
