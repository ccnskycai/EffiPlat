package service

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnvironmentService defines the interface for environment-related business logic.
type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, req models.CreateEnvironmentRequest) (*models.EnvironmentResponse, error)
	GetEnvironments(ctx context.Context, params models.EnvironmentListParams) ([]models.EnvironmentResponse, int64, error)
	GetEnvironmentByID(ctx context.Context, id uint) (*models.EnvironmentResponse, error)
	GetEnvironmentBySlug(ctx context.Context, slug string) (*models.EnvironmentResponse, error)
	UpdateEnvironment(ctx context.Context, id uint, req models.UpdateEnvironmentRequest) (*models.EnvironmentResponse, error)
	DeleteEnvironment(ctx context.Context, id uint) error
}

// ErrEnvironmentNotFound is returned when an environment is not found.
var ErrEnvironmentNotFound = errors.New("environment not found")

// ErrEnvironmentSlugExists is returned when an environment with the same slug already exists.
var ErrEnvironmentSlugExists = errors.New("environment slug already exists")

// ErrEnvironmentNameExists is returned when an environment with the same name already exists.
var ErrEnvironmentNameExists = errors.New("environment name already exists")

type environmentServiceImpl struct {
	repo   repository.EnvironmentRepository
	logger *zap.Logger
}

// NewEnvironmentService creates a new instance of EnvironmentService.
func NewEnvironmentService(repo repository.EnvironmentRepository, logger *zap.Logger) EnvironmentService {
	return &environmentServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *environmentServiceImpl) CreateEnvironment(ctx context.Context, req models.CreateEnvironmentRequest) (*models.EnvironmentResponse, error) {
	s.logger.Info("Service: Creating new environment", zap.String("name", req.Name), zap.String("slug", req.Slug))

	// Check for existing slug
	_, err := s.repo.GetBySlug(ctx, req.Slug)
	if err == nil {
		s.logger.Warn("Service: Environment with this slug already exists", zap.String("slug", req.Slug))
		return nil, ErrEnvironmentSlugExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Service: Error checking for existing slug", zap.String("slug", req.Slug), zap.Error(err))
		return nil, err
	}

	// Optional: Check for existing name if it should be unique across environments not just DB level.
	// For now, relying on DB unique constraint for name if set, or GetByName if implemented.

	env := &models.Environment{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
	}

	createdEnv, err := s.repo.Create(ctx, env)
	if err != nil {
		// TODO: Check for specific DB errors like unique name violation if not checked above
		s.logger.Error("Service: Failed to create environment", zap.Error(err))
		return nil, err
	}

	resp := createdEnv.ToEnvironmentResponse()
	return &resp, nil
}

func (s *environmentServiceImpl) GetEnvironments(ctx context.Context, params models.EnvironmentListParams) ([]models.EnvironmentResponse, int64, error) {
	s.logger.Info("Service: Fetching environments", zap.Any("params", params))
	envs, total, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.Error("Service: Failed to fetch environments", zap.Error(err))
		return nil, 0, err
	}

	responses := make([]models.EnvironmentResponse, len(envs))
	for i, env := range envs {
		responses[i] = env.ToEnvironmentResponse()
	}

	return responses, total, nil
}

func (s *environmentServiceImpl) GetEnvironmentByID(ctx context.Context, id uint) (*models.EnvironmentResponse, error) {
	s.logger.Info("Service: Fetching environment by ID", zap.Uint("id", id))
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Environment not found by ID", zap.Uint("id", id))
			return nil, ErrEnvironmentNotFound
		}
		s.logger.Error("Service: Error fetching environment by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	resp := env.ToEnvironmentResponse()
	return &resp, nil
}

func (s *environmentServiceImpl) GetEnvironmentBySlug(ctx context.Context, slug string) (*models.EnvironmentResponse, error) {
	s.logger.Info("Service: Fetching environment by Slug", zap.String("slug", slug))
	env, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Environment not found by Slug", zap.String("slug", slug))
			return nil, ErrEnvironmentNotFound
		}
		s.logger.Error("Service: Error fetching environment by Slug", zap.String("slug", slug), zap.Error(err))
		return nil, err
	}
	resp := env.ToEnvironmentResponse()
	return &resp, nil
}

func (s *environmentServiceImpl) UpdateEnvironment(ctx context.Context, id uint, req models.UpdateEnvironmentRequest) (*models.EnvironmentResponse, error) {
	s.logger.Info("Service: Updating environment", zap.Uint("id", id))

	existingEnv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Environment not found for update", zap.Uint("id", id))
			return nil, ErrEnvironmentNotFound
		}
		s.logger.Error("Service: Error fetching environment for update", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	updated := false
	if req.Name != nil && *req.Name != existingEnv.Name {
		// Optional: Check if new name conflicts with another existing environment's name
		existingEnv.Name = *req.Name
		updated = true
	}
	if req.Description != nil && *req.Description != existingEnv.Description {
		existingEnv.Description = *req.Description
		updated = true
	}
	if req.Slug != nil && *req.Slug != existingEnv.Slug {
		// Check if new slug conflicts with another existing environment's slug
		foundBySlug, err := s.repo.GetBySlug(ctx, *req.Slug)
		if err == nil && foundBySlug.ID != id { // Slug exists and belongs to another environment
			s.logger.Warn("Service: New slug for update conflicts with existing environment", zap.String("newSlug", *req.Slug), zap.Uint("conflictingEnvID", foundBySlug.ID))
			return nil, ErrEnvironmentSlugExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { // Some other error occurred
			s.logger.Error("Service: Error checking slug for update", zap.String("newSlug", *req.Slug), zap.Error(err))
			return nil, err
		} // If gorm.ErrRecordNotFound, slug is available or it's the same env, proceed.
		existingEnv.Slug = *req.Slug
		updated = true
	}

	if !updated {
		s.logger.Info("Service: No changes detected for environment update", zap.Uint("id", id))
		resp := existingEnv.ToEnvironmentResponse() // Return current state if no updates
		return &resp, nil
	}

	updatedEnv, err := s.repo.Update(ctx, existingEnv)
	if err != nil {
		// Check for specific DB errors
		dbErrStr := err.Error()
		if strings.Contains(dbErrStr, "UNIQUE constraint failed") && strings.Contains(dbErrStr, "environments.name") {
			s.logger.Warn("Service: Environment name conflict during update", zap.Uint("id", id), zap.String("name", existingEnv.Name))
			return nil, ErrEnvironmentNameExists
		}
		// We already check for slug conflict before calling Update
		// if strings.Contains(dbErrStr, "UNIQUE constraint failed") && strings.Contains(dbErrStr, "environments.slug") {
		// 	s.logger.Warn("Service: Environment slug conflict during update", zap.Uint("id", id), zap.String("slug", existingEnv.Slug))
		// 	return nil, ErrEnvironmentSlugExists
		// }
		s.logger.Error("Service: Failed to update environment in repository", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	resp := updatedEnv.ToEnvironmentResponse()
	return &resp, nil
}

func (s *environmentServiceImpl) DeleteEnvironment(ctx context.Context, id uint) error {
	s.logger.Info("Service: Deleting environment", zap.Uint("id", id))
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Environment not found for deletion", zap.Uint("id", id))
			return ErrEnvironmentNotFound
		}
		s.logger.Error("Service: Failed to delete environment", zap.Uint("id", id), zap.Error(err))
		return err
	}
	return nil
}
