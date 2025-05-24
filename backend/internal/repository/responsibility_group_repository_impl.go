package repository

import (
	"EffiPlat/backend/internal/model"
	apputils "EffiPlat/backend/internal/utils"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type gormResponsibilityGroupRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGormResponsibilityGroupRepository creates a new GORM-based ResponsibilityGroupRepository.
func NewGormResponsibilityGroupRepository(db *gorm.DB, logger *zap.Logger) ResponsibilityGroupRepository {
	return &gormResponsibilityGroupRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormResponsibilityGroupRepository) Create(ctx context.Context, group *model.ResponsibilityGroup, responsibilityIDs []uint) (*model.ResponsibilityGroup, error) {
	r.logger.Debug("GORM: Creating responsibility group with responsibilities", zap.String("name", group.Name), zap.Uints("responsibilityIDs", responsibilityIDs))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create the group itself
		if err := tx.Create(group).Error; err != nil {
			r.logger.Error("GORM: Failed to create responsibility group in transaction", zap.Error(err))
			return err // Return error to rollback transaction
		}
		r.logger.Debug("GORM: Group created in transaction, checking ID", zap.Uint("group.ID", group.ID)) // Log ID here

		// 2. If responsibilityIDs are provided, associate them
		if len(responsibilityIDs) > 0 {
			var responsibilitiesToAssociate []model.Responsibility
			for _, respID := range responsibilityIDs {
				// Optional: Validate if respID exists. For now, rely on FK constraints or assume valid IDs.
				responsibilitiesToAssociate = append(responsibilitiesToAssociate, model.Responsibility{ID: respID})
			}
			if err := tx.Model(group).Association("Responsibilities").Replace(responsibilitiesToAssociate); err != nil {
				r.logger.Error("GORM: Failed to associate responsibilities during group creation", zap.Error(err))
				return err // Return error to rollback transaction
			}
		}
		// After successful transaction, GORM would have updated group.ID and potentially other auto-generated fields.
		// To return the group with preloaded responsibilities if they were associated:
		if len(responsibilityIDs) > 0 {
			// Ensure we use the potentially updated group.ID for preloading
			if err := tx.Preload("Responsibilities").First(group, group.ID).Error; err != nil { // group.ID should be set here
				r.logger.Error("GORM: Failed to preload responsibilities after group creation", zap.Uint("groupID_for_preload", group.ID), zap.Error(err))
				return err // Rollback
			}
		}
		return nil // Commit transaction
	})

	if err != nil {
		return nil, err
	}
	r.logger.Debug("GORM: Transaction completed for Create. Returning group.", zap.Uint("final_group.ID", group.ID))
	return group, nil
}

func (r *gormResponsibilityGroupRepository) List(ctx context.Context, params model.ResponsibilityGroupListParams) ([]model.ResponsibilityGroup, int64, error) {
	var groups []model.ResponsibilityGroup
	var total int64
	query := r.db.WithContext(ctx).Model(&model.ResponsibilityGroup{})

	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := 0
	if params.Page > 0 && params.PageSize > 0 {
		offset = (params.Page - 1) * params.PageSize
		query = query.Limit(params.PageSize).Offset(offset)
	} else if params.PageSize > 0 { // Default to page 1 if only PageSize is provided
		query = query.Limit(params.PageSize).Offset(0)
	}

	// Eager load Responsibilities and find records
	err := query.Preload("Responsibilities").Order("created_at DESC").Find(&groups).Error
	return groups, total, err
}

func (r *gormResponsibilityGroupRepository) GetByID(ctx context.Context, id uint) (*model.ResponsibilityGroup, error) {
	r.logger.Debug("GORM: Getting responsibility group by ID", zap.Uint("id", id))
	var group model.ResponsibilityGroup
	// Preload Responsibilities to get associated data
	if err := r.db.WithContext(ctx).Preload("Responsibilities").First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("GORM: Responsibility group not found", zap.Uint("id", id), zap.Error(err))
			return nil, gorm.ErrRecordNotFound // Return gorm.ErrRecordNotFound
		}
		r.logger.Error("GORM: Failed to get responsibility group by ID", zap.Error(err))
		return nil, err
	}
	return &group, nil
}

func (r *gormResponsibilityGroupRepository) Update(ctx context.Context, group *model.ResponsibilityGroup, responsibilityIDs *[]uint) (*model.ResponsibilityGroup, error) {
	r.logger.Debug("GORM: Updating responsibility group", zap.Uint("id", group.ID), zap.Any("responsibilityIDs_ptr", responsibilityIDs != nil))

	if group.ID == 0 {
		return nil, apputils.ErrMissingID
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update the group attributes (Name, Description, etc.)
		// We fetch the group first to ensure it exists and to avoid GORM creating a new one if Save is used on a non-existent ID.
		var existingGroup model.ResponsibilityGroup
		if err := tx.First(&existingGroup, group.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("GORM: Responsibility group not found for update", zap.Uint("id", group.ID), zap.Error(err))
				return gorm.ErrRecordNotFound
			}
			r.logger.Error("GORM: Failed to find group for update", zap.Error(err))
			return err
		}

		// Apply updates to the found group
		existingGroup.Name = group.Name
		existingGroup.Description = group.Description
		// Add other updatable fields here

		if err := tx.Save(&existingGroup).Error; err != nil {
			r.logger.Error("GORM: Failed to save responsibility group updates", zap.Error(err))
			return err
		}
		*group = existingGroup // Reflect updates back to the input `group` pointer for consistency

		// 2. If responsibilityIDs is not nil, update associations
		if responsibilityIDs != nil {
			idsToAssociate := *responsibilityIDs
			// Clear existing associations
			if err := tx.Model(&existingGroup).Association("Responsibilities").Clear(); err != nil {
				r.logger.Error("GORM: Failed to clear responsibilities for group update", zap.Uint("id", existingGroup.ID), zap.Error(err))
				return err
			}

			// Add new associations if any
			if len(idsToAssociate) > 0 {
				var newResponsibilities []model.Responsibility
				for _, respID := range idsToAssociate {
					newResponsibilities = append(newResponsibilities, model.Responsibility{ID: respID})
				}
				if err := tx.Model(&existingGroup).Association("Responsibilities").Append(newResponsibilities); err != nil {
					r.logger.Error("GORM: Failed to append new responsibilities to group during update", zap.Uint("id", existingGroup.ID), zap.Error(err))
					return err
				}
			}
		}

		// 3. Preload responsibilities for the returned group
		if err := tx.Preload("Responsibilities").First(group, group.ID).Error; err != nil {
			r.logger.Error("GORM: Failed to preload responsibilities after group update", zap.Error(err))
			return err
		}
		return nil // Commit transaction
	})

	if err != nil {
		return nil, err
	}
	// `group` is already updated and preloaded by the transaction function
	return group, nil
}

func (r *gormResponsibilityGroupRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Debug("GORM: Deleting responsibility group", zap.Uint("id", id))

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find the group
		var group model.ResponsibilityGroup
		if err := tx.First(&group, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("GORM: Responsibility group not found for deletion", zap.Uint("id", id), zap.Error(err))
				// Consider if not found should be an error or not. For Delete, often it's not.
				// Returning gorm.ErrRecordNotFound for consistency, service layer can decide.
				return gorm.ErrRecordNotFound
			}
			r.logger.Error("GORM: Failed to find responsibility group for deletion", zap.Uint("id", id), zap.Error(err))
			return err
		}

		// Clear the many-to-many associations for "Responsibilities"
		if err := tx.Model(&group).Association("Responsibilities").Clear(); err != nil {
			r.logger.Error("GORM: Failed to clear responsibilities for group", zap.Uint("id", id), zap.Error(err))
			return err
		}

		// Delete the group itself
		if err := tx.Delete(&model.ResponsibilityGroup{}, id).Error; err != nil {
			r.logger.Error("GORM: Failed to delete responsibility group after clearing associations", zap.Uint("id", id), zap.Error(err))
			return err
		}
		return nil
	})
}

func (r *gormResponsibilityGroupRepository) AddResponsibilityToGroup(ctx context.Context, groupID uint, responsibilityID uint) error {
	r.logger.Debug("GORM: Adding responsibility to group", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID))
	association := model.ResponsibilityGroupResponsibility{
		ResponsibilityGroupID: groupID,
		ResponsibilityID:      responsibilityID,
	}
	// Use FirstOrCreate or similar to avoid duplicate errors if the association already exists, or handle the error.
	if err := r.db.WithContext(ctx).Create(&association).Error; err != nil {
		// TODO: Check for specific errors like duplicate entry if primary keys are (groupID, responsibilityID)
		r.logger.Error("GORM: Failed to add responsibility to group", zap.Error(err))
		return err
	}
	return nil
}

func (r *gormResponsibilityGroupRepository) RemoveResponsibilityFromGroup(ctx context.Context, groupID uint, responsibilityID uint) error {
	r.logger.Debug("GORM: Removing responsibility from group", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID))

	result := r.db.WithContext(ctx).Where("responsibility_group_id = ? AND responsibility_id = ?", groupID, responsibilityID).Delete(&model.ResponsibilityGroupResponsibility{})

	if result.Error != nil {
		r.logger.Error("GORM: Failed to remove responsibility from group", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("GORM: Association not found for removal or already removed", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID))
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *gormResponsibilityGroupRepository) GetResponsibilitiesForGroup(ctx context.Context, groupID uint) ([]model.Responsibility, error) {
	r.logger.Debug("GORM: Getting responsibilities for group", zap.Uint("groupID", groupID))
	var group model.ResponsibilityGroup
	if err := r.db.WithContext(ctx).Preload("Responsibilities").First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("GORM: Group not found when getting responsibilities", zap.Uint("groupID", groupID), zap.Error(err))
			return nil, gorm.ErrRecordNotFound // Return gorm.ErrRecordNotFound
		}
		r.logger.Error("GORM: Failed to get group for responsibilities", zap.Error(err))
		return nil, err
	}
	return group.Responsibilities, nil
}

func (r *gormResponsibilityGroupRepository) ReplaceResponsibilitiesForGroup(ctx context.Context, groupID uint, responsibilityIDs []uint) error {
	r.logger.Debug("GORM: Replacing responsibilities for group", zap.Uint("groupID", groupID), zap.Uints("responsibilityIDs", responsibilityIDs))

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group model.ResponsibilityGroup
		if err := tx.First(&group, groupID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("GORM: Group not found for replacing responsibilities", zap.Uint("groupID", groupID), zap.Error(err))
				return gorm.ErrRecordNotFound // Or a specific service error like service.ErrResponsibilityGroupNotFound
			}
			r.logger.Error("GORM: Failed to find group for replacing responsibilities", zap.Error(err))
			return err
		}

		// Clear existing associations
		if err := tx.Model(&group).Association("Responsibilities").Clear(); err != nil {
			r.logger.Error("GORM: Failed to clear existing responsibilities for group", zap.Uint("groupID", groupID), zap.Error(err))
			return err
		}

		// Add new associations if any
		if len(responsibilityIDs) > 0 {
			var newResponsibilities []model.Responsibility
			for _, respID := range responsibilityIDs {
				// Optional: Check if each responsibilityID exists before associating
				// For now, we assume they exist, or DB foreign key constraints will handle it.
				newResponsibilities = append(newResponsibilities, model.Responsibility{ID: respID})
			}
			if err := tx.Model(&group).Association("Responsibilities").Append(newResponsibilities); err != nil {
				r.logger.Error("GORM: Failed to append new responsibilities to group", zap.Uint("groupID", groupID), zap.Error(err))
				return err
			}
		}
		return nil
	})
}
