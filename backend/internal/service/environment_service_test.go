package service_test

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository/mocks"
	"EffiPlat/backend/internal/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Example of a test function for CreateEnvironment
func TestEnvironmentServiceImpl_CreateEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEnvironmentRepository(ctrl)
	logger := zap.NewNop() // Use a Nop logger for tests or a test-specific logger
	s := service.NewEnvironmentService(mockRepo, logger)

	ctx := context.Background()
	createReq := models.CreateEnvironmentRequest{
		Name:        "Test Env",
		Description: "Test Description",
		Slug:        "test-env-slug",
	}

	// Mock successful GetBySlug (slug does not exist)
	mockRepo.EXPECT().GetBySlug(gomock.Eq(ctx), gomock.Eq(createReq.Slug)).Return(nil, gorm.ErrRecordNotFound).Times(1)

	// Mock successful Create
	expectedEnv := &models.Environment{
		ID:          1,
		Name:        createReq.Name,
		Description: createReq.Description,
		Slug:        createReq.Slug,
	}
	mockRepo.EXPECT().Create(gomock.Eq(ctx), gomock.AssignableToTypeOf(&models.Environment{})).Return(expectedEnv, nil).Times(1)

	resp, err := s.CreateEnvironment(ctx, createReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedEnv.Name, resp.Name)
	assert.Equal(t, expectedEnv.Slug, resp.Slug)
	assert.Equal(t, expectedEnv.ID, resp.ID)
}

// Example of a test function for CreateEnvironment when slug already exists
func TestEnvironmentServiceImpl_CreateEnvironment_SlugExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEnvironmentRepository(ctrl)
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
	mockRepo.EXPECT().GetBySlug(gomock.Eq(ctx), gomock.Eq(createReq.Slug)).Return(existingEnv, nil).Times(1)

	resp, err := s.CreateEnvironment(ctx, createReq)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "environment slug 'duplicate-slug' already exists")
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

// Comment block with testify/mock example was here, assuming it's been manually removed or will be.
// If not, it should be deleted as part of standardizing on gomock.
