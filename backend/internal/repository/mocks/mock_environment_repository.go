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
