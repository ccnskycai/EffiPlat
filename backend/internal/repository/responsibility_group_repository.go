//go:generate mockgen -destination=mocks/mock_responsibility_group_repository.go -package=mocks EffiPlat/backend/internal/repository ResponsibilityGroupRepository
package repository

import (
	"EffiPlat/backend/internal/model" // Assuming model.ResponsibilityGroup etc. exist
	"context"
	// "gorm.io/gorm"
)

// ResponsibilityGroupRepository defines the interface for database operations on responsibility groups.
type ResponsibilityGroupRepository interface {
	Create(ctx context.Context, group *model.ResponsibilityGroup, responsibilityIDs []uint) (*model.ResponsibilityGroup, error)
	List(ctx context.Context, params model.ResponsibilityGroupListParams) ([]model.ResponsibilityGroup, int64, error)
	GetByID(ctx context.Context, id uint) (*model.ResponsibilityGroup, error) // Basic group info
	// GetDetailByID(ctx context.Context, id uint) (*model.ResponsibilityGroupDetail, error) // Might include responsibilities
	Update(ctx context.Context, group *model.ResponsibilityGroup, responsibilityIDs *[]uint) (*model.ResponsibilityGroup, error)
	Delete(ctx context.Context, id uint) error

	// Methods for managing the many-to-many relationship with Responsibilities
	AddResponsibilityToGroup(ctx context.Context, groupID uint, responsibilityID uint) error
	RemoveResponsibilityFromGroup(ctx context.Context, groupID uint, responsibilityID uint) error
	ReplaceResponsibilitiesForGroup(ctx context.Context, groupID uint, responsibilityIDs []uint) error // Atomically replace all responsibilities for a group
	GetResponsibilitiesForGroup(ctx context.Context, groupID uint) ([]model.Responsibility, error)
	// RemoveAllResponsibilitiesFromGroup(ctx context.Context, groupID uint) error // Useful for updates
}

/*
// Example GORM implementation structure
type gormResponsibilityGroupRepository struct {
	db *gorm.DB
	logger *zap.Logger
}

func NewGormResponsibilityGroupRepository(db *gorm.DB, logger *zap.Logger) ResponsibilityGroupRepository {
	return &gormResponsibilityGroupRepository{db: db, logger: logger}
}

// ... GORM method implementations ...
// Example for AddResponsibilityToGroup:
// func (r *gormResponsibilityGroupRepository) AddResponsibilityToGroup(ctx context.Context, groupID uint, responsibilityID uint) error {
// 	association := model.ResponsibilityGroupResponsibility{ResponsibilityGroupID: groupID, ResponsibilityID: responsibilityID}
// 	if err := r.db.WithContext(ctx).Create(&association).Error; err != nil {
// 	    // Handle potential duplicate entry errors, etc.
// 	    return err
// 	}
// 	return nil
// }
*/
