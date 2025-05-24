package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"EffiPlat/backend/internal/model" // For all model related structs
	mock_repository "EffiPlat/backend/internal/repository/mocks"
	"EffiPlat/backend/internal/utils" // Renamed from apputils

	// apputils "EffiPlat/backend/internal/utils" // Keep original if specific functions are used from an aliased import

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock" // Import gomock
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Helper function to setup service with mocks using gomock
func newTestServiceInstanceServiceWithMocks(t *testing.T, ctrl *gomock.Controller) (ServiceInstanceService, *mock_repository.MockServiceInstanceRepository, *mock_repository.MockServiceRepository, *mock_repository.MockEnvironmentRepository) {
	mockInstanceRepo := mock_repository.NewMockServiceInstanceRepository(ctrl)
	mockServiceRepo := mock_repository.NewMockServiceRepository(ctrl)
	mockEnvRepo := mock_repository.NewMockEnvironmentRepository(ctrl)
	testLogger := zap.NewNop()

	svc := NewServiceInstanceService(mockInstanceRepo, mockServiceRepo, mockEnvRepo, testLogger)
	return svc, mockInstanceRepo, mockServiceRepo, mockEnvRepo
}

func TestServiceInstanceServiceImpl_CreateServiceInstance(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, mockInstanceRepo, mockServiceRepo, mockEnvRepo := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()

		inputDTO := &ServiceInstanceInputDTO{
			ServiceID:     1,
			EnvironmentID: 1,
			Version:       "1.0.0",
			Status:        string(model.ServiceInstanceStatusDeploying),
			Config:        datatypes.JSONMap{"key": "value"},
		}
		expectedService := &model.Service{ID: inputDTO.ServiceID}
		expectedEnv := &model.Environment{ID: inputDTO.EnvironmentID}

		// Mock dependencies
		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(expectedService, nil)
		mockEnvRepo.EXPECT().GetByID(ctx, inputDTO.EnvironmentID).Return(expectedEnv, nil)
		mockInstanceRepo.EXPECT().CheckExists(ctx, inputDTO.ServiceID, inputDTO.EnvironmentID, inputDTO.Version, uint(0)).Return(false, nil)

		// Mock the Create call on instanceRepo
		mockInstanceRepo.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&model.ServiceInstance{})).
			DoAndReturn(func(_ context.Context, si *model.ServiceInstance) error {
				si.ID = 123 // Simulate GORM setting the ID
				si.CreatedAt = time.Now()
				si.UpdatedAt = time.Now()
				// Match DTO fields to the model being created
				assert.Equal(t, inputDTO.ServiceID, si.ServiceID)
				assert.Equal(t, inputDTO.EnvironmentID, si.EnvironmentID)
				assert.Equal(t, inputDTO.Version, si.Version)
				assert.Equal(t, model.ServiceInstanceStatusType(inputDTO.Status), si.Status)
				// Config needs careful handling if it's a custom type or interface
				assert.Equal(t, inputDTO.Config, datatypes.JSONMap(si.Config)) // Assuming input DTO Config matches model Config type
				return nil
			})

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.NoError(t, err)
		assert.NotNil(t, outputDTO)
		assert.Equal(t, uint(123), outputDTO.ID)
		assert.Equal(t, inputDTO.ServiceID, outputDTO.ServiceID)
		assert.Equal(t, inputDTO.Version, outputDTO.Version)
		assert.Equal(t, inputDTO.Status, outputDTO.Status)
		assert.Equal(t, inputDTO.Config, outputDTO.Config)
		// Assertions for CreatedAt and UpdatedAt can be tricky with time.Now().
		// Consider checking if they are non-zero or within a small duration if exact match is not possible.
		assert.NotZero(t, outputDTO.CreatedAt)
		assert.NotZero(t, outputDTO.UpdatedAt)
	})

	t.Run("Create - Service not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, _, mockServiceRepo, _ := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()
		inputDTO := &ServiceInstanceInputDTO{ServiceID: 99}

		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(nil, gorm.ErrRecordNotFound)

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.Error(t, err)
		assert.Nil(t, outputDTO)
		assert.True(t, errors.Is(err, utils.ErrBadRequest), "expected ErrBadRequest")
		assert.Contains(t, err.Error(), fmt.Sprintf("service with ID %d not found", inputDTO.ServiceID))
	})

	t.Run("Create - Environment not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, _, mockServiceRepo, mockEnvRepo := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()
		inputDTO := &ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 99}
		expectedService := &model.Service{ID: inputDTO.ServiceID}

		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(expectedService, nil)
		mockEnvRepo.EXPECT().GetByID(ctx, inputDTO.EnvironmentID).Return(nil, gorm.ErrRecordNotFound)

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.Error(t, err)
		assert.Nil(t, outputDTO)
		assert.True(t, errors.Is(err, utils.ErrBadRequest), "expected ErrBadRequest")
		assert.Contains(t, err.Error(), fmt.Sprintf("environment with ID %d not found", inputDTO.EnvironmentID))
	})

	t.Run("Create - Instance already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, mockInstanceRepo, mockServiceRepo, mockEnvRepo := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()
		inputDTO := &ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "1.0.0"}
		expectedService := &model.Service{ID: inputDTO.ServiceID}
		expectedEnv := &model.Environment{ID: inputDTO.EnvironmentID}

		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(expectedService, nil)
		mockEnvRepo.EXPECT().GetByID(ctx, inputDTO.EnvironmentID).Return(expectedEnv, nil)
		mockInstanceRepo.EXPECT().CheckExists(ctx, inputDTO.ServiceID, inputDTO.EnvironmentID, inputDTO.Version, uint(0)).Return(true, nil)

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.Error(t, err)
		assert.Nil(t, outputDTO)
		assert.True(t, errors.Is(err, utils.ErrAlreadyExists), "expected ErrAlreadyExists")
	})

	t.Run("Create - CheckExists fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, mockInstanceRepo, mockServiceRepo, mockEnvRepo := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()
		dbError := errors.New("db check error")
		inputDTO := &ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "1.0.0"}
		expectedService := &model.Service{ID: inputDTO.ServiceID}
		expectedEnv := &model.Environment{ID: inputDTO.EnvironmentID}

		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(expectedService, nil)
		mockEnvRepo.EXPECT().GetByID(ctx, inputDTO.EnvironmentID).Return(expectedEnv, nil)
		mockInstanceRepo.EXPECT().CheckExists(ctx, inputDTO.ServiceID, inputDTO.EnvironmentID, inputDTO.Version, uint(0)).Return(false, dbError)

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.Error(t, err)
		assert.Nil(t, outputDTO)
		// Consider wrapping the error in service layer if not already
		assert.Contains(t, err.Error(), "failed to check for existing instance")
	})

	t.Run("Create - Repository create fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc, mockInstanceRepo, mockServiceRepo, mockEnvRepo := newTestServiceInstanceServiceWithMocks(t, ctrl)
		ctx := context.Background()
		dbError := errors.New("db create error")
		inputDTO := &ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "1.0.0"}
		expectedService := &model.Service{ID: inputDTO.ServiceID}
		expectedEnv := &model.Environment{ID: inputDTO.EnvironmentID}

		mockServiceRepo.EXPECT().GetByID(ctx, inputDTO.ServiceID).Return(expectedService, nil)
		mockEnvRepo.EXPECT().GetByID(ctx, inputDTO.EnvironmentID).Return(expectedEnv, nil)
		mockInstanceRepo.EXPECT().CheckExists(ctx, inputDTO.ServiceID, inputDTO.EnvironmentID, inputDTO.Version, uint(0)).Return(false, nil)
		mockInstanceRepo.EXPECT().Create(ctx, gomock.AssignableToTypeOf(&model.ServiceInstance{})).Return(dbError)

		outputDTO, err := svc.CreateServiceInstance(ctx, inputDTO)

		assert.Error(t, err)
		assert.Nil(t, outputDTO)
		// Consider wrapping the error in service layer
		assert.Contains(t, err.Error(), "failed to create service instance")
	})
}

// TODO: Add tests for GetServiceInstanceByID, ListServiceInstances, UpdateServiceInstance, DeleteServiceInstance
// using gomock patterns.
