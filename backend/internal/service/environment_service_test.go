package service_test

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository/mocks" // Assuming you'll use mocks for the repository
	"EffiPlat/backend/internal/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Example of a test function for CreateEnvironment
func TestEnvironmentServiceImpl_CreateEnvironment(t *testing.T) {
	mockRepo := new(mocks.MockEnvironmentRepository)
	logger := zap.NewNop() // Use a Nop logger for tests or a test-specific logger
	s := service.NewEnvironmentService(mockRepo, logger)

	ctx := context.Background()
	createReq := models.CreateEnvironmentRequest{
		Name:        "Test Env",
		Description: "Test Description",
		Slug:        "test-env-slug",
	}

	// Mock successful GetBySlug (slug does not exist)
	mockRepo.On("GetBySlug", ctx, createReq.Slug).Return(nil, gorm.ErrRecordNotFound).Once()

	// Mock successful Create
	expectedEnv := &models.Environment{
		ID:          1,
		Name:        createReq.Name,
		Description: createReq.Description,
		Slug:        createReq.Slug,
	}
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Environment")).Return(expectedEnv, nil).Once()

	resp, err := s.CreateEnvironment(ctx, createReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedEnv.Name, resp.Name)
	assert.Equal(t, expectedEnv.Slug, resp.Slug)
	assert.Equal(t, expectedEnv.ID, resp.ID)

	mockRepo.AssertExpectations(t)
}

// Example of a test function for CreateEnvironment when slug already exists
func TestEnvironmentServiceImpl_CreateEnvironment_SlugExists(t *testing.T) {
	mockRepo := new(mocks.MockEnvironmentRepository)
	logger := zap.NewNop()
	s := service.NewEnvironmentService(mockRepo, logger)

	ctx := context.Background()
	createReq := models.CreateEnvironmentRequest{
		Name:        "Test Env Duplicate Slug",
		Description: "Test Description",
		Slug:        "duplicate-slug",
	}

	// Mock GetBySlug finding an existing environment
	existingEnv := &models.Environment{
		ID:   2,
		Name: "Some Other Env",
		Slug: "duplicate-slug",
	}
	mockRepo.On("GetBySlug", ctx, createReq.Slug).Return(existingEnv, nil).Once()

	resp, err := s.CreateEnvironment(ctx, createReq)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "environment slug 'duplicate-slug' already exists")

	mockRepo.AssertExpectations(t)
}

// You would continue to add more test functions for other methods like:
// - TestEnvironmentServiceImpl_GetEnvironments
// - TestEnvironmentServiceImpl_GetEnvironmentByID
// - TestEnvironmentServiceImpl_GetEnvironmentByID_NotFound
// - TestEnvironmentServiceImpl_GetEnvironmentBySlug
// - TestEnvironmentServiceImpl_GetEnvironmentBySlug_NotFound
// - TestEnvironmentServiceImpl_UpdateEnvironment
// - TestEnvironmentServiceImpl_UpdateEnvironment_NotFound
// - TestEnvironmentServiceImpl_UpdateEnvironment_SlugConflict
// - TestEnvironmentServiceImpl_UpdateEnvironment_NameConflict
// - TestEnvironmentServiceImpl_DeleteEnvironment
// - TestEnvironmentServiceImpl_DeleteEnvironment_NotFound

// Remember to create mock implementations for your repository dependencies.
// For example, if you are using testify/mock, you might have a file like:
// /Users/skyccn/EffiPlat/backend/internal/repository/mocks/mock_environment_repository.go
// with content similar to:
/*
package mocks

import (
	"EffiPlat/backend/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	args := m.Called(ctx, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) List(ctx context.Context, params models.EnvironmentListParams) ([]models.Environment, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Environment), args.Get(1).(int64), args.Error(2)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id uint) (*models.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetBySlug(ctx context.Context, slug string) (*models.Environment, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, environment *models.Environment) (*models.Environment, error) {
	args := m.Called(ctx, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
*/