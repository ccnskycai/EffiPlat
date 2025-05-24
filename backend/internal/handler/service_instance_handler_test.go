package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/service"
	mock_service "EffiPlat/backend/internal/service/mocks"
	apputils "EffiPlat/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

func setupRouterAndHandler(t *testing.T) (*gin.Engine, *handler.ServiceInstanceHandler, *mock_service.MockServiceInstanceService, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	logger, _ := zap.NewDevelopment()

	mockSvc := mock_service.NewMockServiceInstanceService(ctrl)
	instanceHandler := handler.NewServiceInstanceHandler(mockSvc, logger)

	return router, instanceHandler, mockSvc, ctrl
}

func TestServiceInstanceHandler_CreateServiceInstance(t *testing.T) {
	router, instanceHandler, mockSvc, ctrl := setupRouterAndHandler(t)
	defer ctrl.Finish()

	// 注册路由
	serviceInstanceGroup := router.Group("/api/v1/service-instances")
	{
		serviceInstanceGroup.POST("", instanceHandler.CreateServiceInstance)
	}

	t.Run("Successful creation", func(t *testing.T) {
		inputDTO := service.ServiceInstanceInputDTO{
			ServiceID:     1,
			EnvironmentID: 1,
			Version:       "1.0.0",
			Status:        string(model.ServiceInstanceStatusDeploying),
			Config:        datatypes.JSONMap{"key": "value"},
		}
		expectedOutputDTO := &service.ServiceInstanceOutputDTO{
			ID:            123,
			ServiceID:     inputDTO.ServiceID,
			EnvironmentID: inputDTO.EnvironmentID,
			Version:       inputDTO.Version,
			Status:        inputDTO.Status,
			Config:        inputDTO.Config,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mockSvc.EXPECT().CreateServiceInstance(gomock.Any(), gomock.Any()).Return(expectedOutputDTO, nil).Times(1)

		jsonBody, _ := json.Marshal(inputDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-instances", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		// Define a wrapper struct to match the actual JSON structure
		type SuccessResponseWrapper struct {
			Code    int                              `json:"code"`
			Message string                           `json:"message"`
			Data    service.ServiceInstanceOutputDTO `json:"data"` // Note: Create returns *DTO, but JSON is value
		}
		var responseWrapper SuccessResponseWrapper
		err := json.Unmarshal(rr.Body.Bytes(), &responseWrapper)
		assert.NoError(t, err)

		actualResponse := responseWrapper.Data // Extract the actual DTO
		assert.Equal(t, expectedOutputDTO.ID, actualResponse.ID)
		assert.Equal(t, expectedOutputDTO.Version, actualResponse.Version)
	})

	t.Run("Invalid request payload - Bind error", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-instances", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse.Message, "Invalid request payload")
	})

	t.Run("Service layer returns ErrBadRequest", func(t *testing.T) {
		inputDTO := service.ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "v1", Status: "running"}
		serviceError := fmt.Errorf("%w: specific validation error", apputils.ErrBadRequest)
		mockSvc.EXPECT().CreateServiceInstance(gomock.Any(), gomock.Eq(&inputDTO)).Return(nil, serviceError).Times(1)

		jsonBody, _ := json.Marshal(inputDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-instances", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, serviceError.Error(), errorResponse.Message)
	})

	t.Run("Service layer returns ErrAlreadyExists", func(t *testing.T) {
		inputDTO := service.ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "v1", Status: "running"}
		serviceError := fmt.Errorf("%w: instance already exists", apputils.ErrAlreadyExists)
		mockSvc.EXPECT().CreateServiceInstance(gomock.Any(), gomock.Eq(&inputDTO)).Return(nil, serviceError).Times(1)

		jsonBody, _ := json.Marshal(inputDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-instances", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)

		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, serviceError.Error(), errorResponse.Message)
	})

	t.Run("Service layer returns generic error", func(t *testing.T) {
		inputDTO := service.ServiceInstanceInputDTO{ServiceID: 1, EnvironmentID: 1, Version: "v1", Status: "running"}
		serviceError := errors.New("some unexpected internal error")
		mockSvc.EXPECT().CreateServiceInstance(gomock.Any(), gomock.Eq(&inputDTO)).Return(nil, serviceError).Times(1)

		jsonBody, _ := json.Marshal(inputDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/service-instances", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to create service instance", errorResponse.Message)
	})
}

func TestServiceInstanceHandler_GetServiceInstance(t *testing.T) {
	router, instanceHandler, mockSvc, ctrl := setupRouterAndHandler(t)
	defer ctrl.Finish()

	serviceInstanceGroup := router.Group("/api/v1/service-instances")
	{
		serviceInstanceGroup.GET("/:instanceId", instanceHandler.GetServiceInstance)
	}

	t.Run("Successful get", func(t *testing.T) {
		instanceID := uint(123)
		expectedOutputDTO := &service.ServiceInstanceOutputDTO{
			ID:            instanceID,
			ServiceID:     1,
			EnvironmentID: 1,
			Version:       "1.0.0",
			Status:        string(model.ServiceInstanceStatusRunning),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mockSvc.EXPECT().GetServiceInstanceByID(gomock.Any(), gomock.Eq(instanceID)).Return(expectedOutputDTO, nil).Times(1)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/service-instances/%d", instanceID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		t.Logf("Raw JSON response: %s", rr.Body.String())

		// Define a wrapper struct to match the actual JSON structure
		type SuccessResponseWrapper struct {
			Code    int                              `json:"code"`
			Message string                           `json:"message"`
			Data    service.ServiceInstanceOutputDTO `json:"data"`
		}
		var responseWrapper SuccessResponseWrapper
		err := json.Unmarshal(rr.Body.Bytes(), &responseWrapper)
		assert.NoError(t, err)

		actualResponse := responseWrapper.Data // Extract the actual DTO
		assert.Equal(t, expectedOutputDTO.ID, actualResponse.ID)
		assert.Equal(t, expectedOutputDTO.Version, actualResponse.Version)
	})

	t.Run("Invalid instance ID format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/service-instances/invalid-id", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid instance ID format", errorResponse.Message)
	})

	t.Run("Service layer returns ErrNotFound", func(t *testing.T) {
		instanceID := uint(404)
		serviceError := apputils.ErrNotFound
		mockSvc.EXPECT().GetServiceInstanceByID(gomock.Any(), gomock.Eq(instanceID)).Return(nil, serviceError).Times(1)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/service-instances/%d", instanceID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Service instance not found", errorResponse.Message)
	})

	t.Run("Service layer returns generic error", func(t *testing.T) {
		instanceID := uint(500)
		serviceError := errors.New("some generic error")
		mockSvc.EXPECT().GetServiceInstanceByID(gomock.Any(), gomock.Eq(instanceID)).Return(nil, serviceError).Times(1)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/service-instances/%d", instanceID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		var errorResponse models.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, "Failed to retrieve service instance", errorResponse.Message)
	})
}
