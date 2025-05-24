package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BusinessInputDTO 定义了创建和更新业务时的输入数据结构
type BusinessInputDTO struct {
	Name        *string                   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string                   `json:"description,omitempty" validate:"omitempty,max=500"`
	Owner       *string                   `json:"owner,omitempty" validate:"omitempty,email,max=100"`
	Status      *model.BusinessStatusType `json:"status,omitempty" validate:"omitempty,is_business_status"`
}

// BusinessOutputDTO 定义了展示业务数据时的输出结构
type BusinessOutputDTO struct {
	ID          uint                     `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Owner       string                   `json:"owner,omitempty"`
	Status      model.BusinessStatusType `json:"status"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

// ListBusinessesParamsDTO 定义了列出业务时的查询参数，供 handler 使用
// 它将映射到 repository.ListBusinessesParams
type ListBusinessesParamsDTO struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Name     string `form:"name"`
	Status   string `form:"status" validate:"omitempty,is_business_status_string"`
	Owner    string `form:"owner"`
	SortBy   string `form:"sortBy,default=created_at"`
	Order    string `form:"order,default=desc" validate:"omitempty,oneof=asc desc"`
}

// BusinessService 定义业务逻辑服务接口
type BusinessService interface {
	CreateBusiness(ctx context.Context, input BusinessInputDTO) (*BusinessOutputDTO, error)
	GetBusinessByID(ctx context.Context, id uint) (*BusinessOutputDTO, error)
	ListBusinesses(ctx context.Context, params ListBusinessesParamsDTO) ([]BusinessOutputDTO, int64, error)
	UpdateBusiness(ctx context.Context, id uint, input BusinessInputDTO) (*BusinessOutputDTO, error)
	DeleteBusiness(ctx context.Context, id uint) error
}

type businessServiceImpl struct {
	repo   repository.BusinessRepository
	logger *zap.Logger
}

// NewBusinessService 创建一个新的 BusinessService 实例
func NewBusinessService(repo repository.BusinessRepository, logger *zap.Logger) BusinessService {
	// 注册自定义验证器
	utils.GetValidator().RegisterValidation("is_business_status", IsBusinessStatusValid)
	utils.GetValidator().RegisterValidation("is_business_status_string", IsBusinessStatusStringValid)
	return &businessServiceImpl{repo: repo, logger: logger}
}

// mapModelToOutputDTO 将 model.Business 转换为 BusinessOutputDTO
func mapModelToOutputDTO(business *model.Business) *BusinessOutputDTO {
	if business == nil {
		return nil
	}
	return &BusinessOutputDTO{
		ID:          business.ID,
		Name:        business.Name,
		Description: business.Description,
		Owner:       business.Owner,
		Status:      business.Status,
		CreatedAt:   business.CreatedAt,
		UpdatedAt:   business.UpdatedAt,
	}
}

// CreateBusiness 创建一个新的业务
func (s *businessServiceImpl) CreateBusiness(ctx context.Context, input BusinessInputDTO) (*BusinessOutputDTO, error) {
	s.logger.Debug("Service: Attempting to create business", zap.Any("input", input))

	// Explicit check for Name on create, as it's now a pointer in DTO
	if input.Name == nil || *input.Name == "" {
		s.logger.Warn("Service: CreateBusiness validation failed - Name is required")
		return nil, fmt.Errorf("%w: Name is required", utils.ErrBadRequest)
	}

	if err := utils.GetValidator().Struct(input); err != nil {
		s.logger.Warn("Service: CreateBusiness validation failed", zap.Error(err))
		// Enhance error message to be more specific if possible from validator
		return nil, fmt.Errorf("%w: %s", utils.ErrBadRequest, utils.FormatValidationError(err))
	}

	exists, err := s.repo.CheckExists(ctx, *input.Name, 0) // Use *input.Name
	if err != nil {
		s.logger.Error("Service: Error checking business existence", zap.Error(err))
		return nil, fmt.Errorf("service.CreateBusiness.CheckExists: %w", err)
	}
	if exists {
		s.logger.Warn("Service: Business with this name already exists", zap.String("name", *input.Name))
		return nil, fmt.Errorf("%w: business with name '%s' already exists", utils.ErrAlreadyExists, *input.Name)
	}

	business := &model.Business{
		Name: *input.Name, // Dereference Name
	}
	if input.Description != nil {
		business.Description = *input.Description // Dereference Description
	}
	if input.Owner != nil {
		business.Owner = *input.Owner // Dereference Owner
	}

	if input.Status != nil {
		business.Status = *input.Status
	} else {
		business.Status = model.BusinessStatusActive // 默认状态
	}

	if err := s.repo.Create(ctx, business); err != nil {
		s.logger.Error("Service: Failed to create business in repository", zap.Error(err))
		return nil, fmt.Errorf("service.CreateBusiness.RepoCreate: %w", err)
	}

	s.logger.Info("Service: Business created successfully", zap.Uint("id", business.ID))
	return mapModelToOutputDTO(business), nil
}

// GetBusinessByID 通过 ID 获取业务信息
func (s *businessServiceImpl) GetBusinessByID(ctx context.Context, id uint) (*BusinessOutputDTO, error) {
	s.logger.Debug("Service: Getting business by ID", zap.Uint("id", id))
	if id == 0 {
		return nil, fmt.Errorf("%w: business ID cannot be zero", utils.ErrBadRequest)
	}

	business, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Business not found by ID", zap.Uint("id", id))
			return nil, utils.ErrNotFound
		}
		s.logger.Error("Service: Failed to get business by ID from repository", zap.Error(err))
		return nil, fmt.Errorf("service.GetBusinessByID.RepoGet: %w", err)
	}
	return mapModelToOutputDTO(business), nil
}

// ListBusinesses 列出业务，支持过滤和分页
func (s *businessServiceImpl) ListBusinesses(ctx context.Context, paramsDTO ListBusinessesParamsDTO) ([]BusinessOutputDTO, int64, error) {
	s.logger.Debug("Service: Listing businesses", zap.Any("paramsDTO", paramsDTO))
	if err := utils.GetValidator().Struct(paramsDTO); err != nil {
		s.logger.Warn("Service: ListBusinesses validation failed", zap.Error(err))
		return nil, 0, fmt.Errorf("%w: %s", utils.ErrBadRequest, err.Error())
	}

	// Map DTO sortBy to actual database column names
	dbSortBy := paramsDTO.SortBy
	switch paramsDTO.SortBy {
	case "createdAt":
		dbSortBy = "created_at"
	case "updatedAt":
		dbSortBy = "updated_at"
	case "name":
		dbSortBy = "name"
	case "status":
		dbSortBy = "status"
	case "owner":
		dbSortBy = "owner"
		// default to paramsDTO.SortBy if not a recognized alias,
		// or if it's already a correct column name (e.g. "created_at" passed directly)
	}
	if dbSortBy == "" { // Default sort if not provided or mapped to empty
		dbSortBy = "created_at" // Default to created_at if empty after mapping
	}

	repoParams := repository.ListBusinessesParams{
		Page:     paramsDTO.Page,
		PageSize: paramsDTO.PageSize,
		SortBy:   dbSortBy, // Use the mapped dbSortBy
		Order:    paramsDTO.Order,
	}
	if paramsDTO.Name != "" {
		repoParams.Name = &paramsDTO.Name
	}
	if paramsDTO.Owner != "" {
		repoParams.Owner = &paramsDTO.Owner
	}
	if paramsDTO.Status != "" {
		statusEnum := model.BusinessStatusType(paramsDTO.Status)
		repoParams.Status = &statusEnum
	}

	browsers, total, err := s.repo.List(ctx, &repoParams)
	if err != nil {
		s.logger.Error("Service: Failed to list businesses from repository", zap.Error(err))
		return nil, 0, fmt.Errorf("service.ListBusinesses.RepoList: %w", err)
	}

	outputDTOs := make([]BusinessOutputDTO, len(browsers))
	for i, b := range browsers {
		outputDTOs[i] = *mapModelToOutputDTO(&b)
	}

	return outputDTOs, total, nil
}

// UpdateBusiness 更新现有业务
func (s *businessServiceImpl) UpdateBusiness(ctx context.Context, id uint, input BusinessInputDTO) (*BusinessOutputDTO, error) {
	s.logger.Debug("Service: Attempting to update business", zap.Uint("id", id), zap.Any("input", input))
	if id == 0 {
		return nil, fmt.Errorf("%w: business ID cannot be zero for update", utils.ErrBadRequest)
	}

	// Validate the input DTO. Note: omitempty on fields means they are optional.
	// Validation rules (like min=2 for Name) will only apply if the field is provided.
	if err := utils.GetValidator().Struct(input); err != nil {
		s.logger.Warn("Service: UpdateBusiness validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %s", utils.ErrBadRequest, utils.FormatValidationError(err))
	}

	existingBusiness, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Business not found for update", zap.Uint("id", id))
			return nil, utils.ErrNotFound
		}
		s.logger.Error("Service: Failed to get business for update from repository", zap.Error(err))
		return nil, fmt.Errorf("service.UpdateBusiness.RepoGet: %w", err)
	}

	needsUpdate := false
	if input.Name != nil {
		if *input.Name != existingBusiness.Name { // Check if new name is different
			// Check for name conflict only if name is actually changing to a new value
			exists, err := s.repo.CheckExists(ctx, *input.Name, id)
			if err != nil {
				s.logger.Error("Service: Error checking new business name existence", zap.Error(err))
				return nil, fmt.Errorf("service.UpdateBusiness.CheckExists: %w", err)
			}
			if exists {
				s.logger.Warn("Service: Another business with this new name already exists", zap.String("newName", *input.Name))
				return nil, fmt.Errorf("%w: business with name '%s' already exists", utils.ErrAlreadyExists, *input.Name)
			}
			existingBusiness.Name = *input.Name
			needsUpdate = true
		}
	}

	if input.Description != nil {
		if *input.Description != existingBusiness.Description {
			existingBusiness.Description = *input.Description
			needsUpdate = true
		}
	}

	if input.Owner != nil {
		if *input.Owner != existingBusiness.Owner {
			existingBusiness.Owner = *input.Owner
			needsUpdate = true
		}
	}

	if input.Status != nil {
		if *input.Status != existingBusiness.Status {
			existingBusiness.Status = *input.Status
			needsUpdate = true
		}
	}

	if !needsUpdate {
		s.logger.Info("Service: No actual changes detected for business update", zap.Uint("id", id))
		return mapModelToOutputDTO(existingBusiness), nil // No changes, return current state
	}

	if err := s.repo.Update(ctx, existingBusiness); err != nil {
		s.logger.Error("Service: Failed to update business in repository", zap.Error(err))
		// 根据 repository.Update 的行为，它可能在记录未找到时也返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) { // 以防万一，repository没处理好
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("service.UpdateBusiness.RepoUpdate: %w", err)
	}

	s.logger.Info("Service: Business updated successfully", zap.Uint("id", existingBusiness.ID))
	return mapModelToOutputDTO(existingBusiness), nil
}

// DeleteBusiness 删除一个业务
func (s *businessServiceImpl) DeleteBusiness(ctx context.Context, id uint) error {
	s.logger.Debug("Service: Deleting business", zap.Uint("id", id))
	if id == 0 {
		return fmt.Errorf("%w: business ID cannot be zero for delete", utils.ErrBadRequest)
	}

	// 检查业务是否存在，然后删除，确保返回正确的错误类型
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Business not found for deletion", zap.Uint("id", id))
			return utils.ErrNotFound
		}
		s.logger.Error("Service: Failed to get business for deletion check", zap.Error(err))
		return fmt.Errorf("service.DeleteBusiness.RepoGet: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Service: Failed to delete business in repository", zap.Error(err))
		// repo.Delete 应该在记录未找到时返回 gorm.ErrRecordNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return fmt.Errorf("service.DeleteBusiness.RepoDelete: %w", err)
	}

	s.logger.Info("Service: Business deleted successfully", zap.Uint("id", id))
	return nil
}

// IsBusinessStatusValid 是一个自定义验证函数，用于验证 model.BusinessStatusType
func IsBusinessStatusValid(fl validator.FieldLevel) bool {
	if status, ok := fl.Field().Interface().(model.BusinessStatusType); ok {
		return status.IsValid()
	}
	// 如果是指针类型，则需要解引用
	if pStatus, ok := fl.Field().Interface().(*model.BusinessStatusType); ok {
		if pStatus != nil {
			return (*pStatus).IsValid()
		}
		return true // nil 指针被认为是有效的 (即，不更新)
	}
	return false
}

// IsBusinessStatusStringValid 是一个自定义验证函数，用于验证字符串形式的业务状态
func IsBusinessStatusStringValid(fl validator.FieldLevel) bool {
	statusStr := fl.Field().String()
	if statusStr == "" { // 允许为空，表示不按状态过滤
		return true
	}
	status := model.BusinessStatusType(statusStr)
	return status.IsValid()
}
