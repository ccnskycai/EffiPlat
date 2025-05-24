package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBugRepository is a mock implementation of repository.BugRepository for testing
type MockBugRepository struct {
	mock.Mock
}

func (m *MockBugRepository) Create(ctx context.Context, bug *model.Bug) error {
	args := m.Called(ctx, bug)
	return args.Error(0)
}

func (m *MockBugRepository) GetByID(ctx context.Context, id uint) (*model.Bug, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Bug), args.Error(1)
}

func (m *MockBugRepository) Update(ctx context.Context, bug *model.Bug) error {
	args := m.Called(ctx, bug)
	return args.Error(0)
}

func (m *MockBugRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBugRepository) List(ctx context.Context, params *model.BugListParams) ([]*model.Bug, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*model.Bug), args.Get(1).(int64), args.Error(2)
}

func (m *MockBugRepository) CountBugsByAssigneeID(ctx context.Context, assigneeID uint) (int64, error) {
	args := m.Called(ctx, assigneeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBugRepository) CountBugsByEnvironmentID(ctx context.Context, environmentID uint) (int64, error) {
	args := m.Called(ctx, environmentID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBugRepository) GetBugsByStatus(ctx context.Context, status model.BugStatusType, params *model.BugListParams) ([]*model.Bug, int64, error) {
	args := m.Called(ctx, status, params)
	return args.Get(0).([]*model.Bug), args.Get(1).(int64), args.Error(2)
}

// Helper function to setup the service and mock repository
func setupBugServiceTest(t *testing.T) (BugService, *MockBugRepository) {
	mockRepo := new(MockBugRepository)
	service := NewBugService(mockRepo)
	return service, mockRepo
}

// TestBugService_CreateBug tests the CreateBug method
func TestBugService_CreateBug(t *testing.T) {
	service, mockRepo := setupBugServiceTest(t)
	ctx := context.Background()

	// Test case: Successful bug creation
	t.Run("Successful Creation", func(t *testing.T) {
		req := &model.CreateBugRequest{
			Title:       "Test Bug",
			Description: "This is a test bug",
			Status:      model.BugStatusOpen,
			Priority:    model.BugPriorityMedium,
		}

		// Setting up the mock to modify the bug passed to Create (to set its ID)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Bug")).Run(func(args mock.Arguments) {
			bug := args.Get(1).(*model.Bug)
			bug.ID = 1 // Simulate ID assignment
			bug.CreatedAt = time.Now()
			bug.UpdatedAt = time.Now()
		}).Return(nil).Once()

		// Call the service method
		resp, err := service.CreateBug(ctx, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Title, resp.Title)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Repository error
	t.Run("Repository Error", func(t *testing.T) {
		req := &model.CreateBugRequest{
			Title:       "Test Bug",
			Description: "This is a test bug",
			Status:      model.BugStatusOpen,
			Priority:    model.BugPriorityMedium,
		}

		// Need to reset the mock between test cases
		mockRepo := new(MockBugRepository)
		service := NewBugService(mockRepo)

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Bug")).Return(errors.New("database error")).Once()

		resp, err := service.CreateBug(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

// TestBugService_GetBugByID tests the GetBugByID method
func TestBugService_GetBugByID(t *testing.T) {
	service, mockRepo := setupBugServiceTest(t)
	ctx := context.Background()

	// Test case: Bug found
	t.Run("Bug Found", func(t *testing.T) {
		bugID := uint(1)
		mockBug := &model.Bug{
			ID:          bugID,
			Title:       "Existing Bug",
			Description: "This bug exists",
			Status:      model.BugStatusOpen,
			Priority:    model.BugPriorityHigh,
		}

		mockRepo.On("GetByID", ctx, bugID).Return(mockBug, nil).Once()

		resp, err := service.GetBugByID(ctx, bugID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, mockBug.Title, resp.Title)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Bug not found
	t.Run("Bug Not Found", func(t *testing.T) {
		bugID := uint(99)
		mockRepo.On("GetByID", ctx, bugID).Return(nil, repository.ErrBugNotFound).Once()

		resp, err := service.GetBugByID(ctx, bugID)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, repository.ErrBugNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestBugService_UpdateBug tests the UpdateBug method
func TestBugService_UpdateBug(t *testing.T) {
	service, mockRepo := setupBugServiceTest(t)
	ctx := context.Background()

	// Test case: Successful update
	t.Run("Successful Update", func(t *testing.T) {
		bugID := uint(1)
		existingBug := &model.Bug{
			ID:          bugID,
			Title:       "Existing Bug",
			Description: "Original description",
			Status:      model.BugStatusOpen,
			Priority:    model.BugPriorityMedium,
		}

		updateReq := &model.UpdateBugRequest{
			Title:    strPtr("Updated Bug Title"),
			Status:   statusPtr(model.BugStatusInProgress),
			Priority: priorityPtr(model.BugPriorityHigh),
		}

		mockRepo.On("GetByID", ctx, bugID).Return(existingBug, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Bug")).Return(nil).Once()

		resp, err := service.UpdateBug(ctx, bugID, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, *updateReq.Title, resp.Title)
		assert.Equal(t, *updateReq.Status, resp.Status)
		assert.Equal(t, *updateReq.Priority, resp.Priority)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Bug not found
	t.Run("Bug Not Found", func(t *testing.T) {
		bugID := uint(99)
		updateReq := &model.UpdateBugRequest{
			Title: strPtr("Updated Bug Title"),
		}

		mockRepo.On("GetByID", ctx, bugID).Return(nil, repository.ErrBugNotFound).Once()

		resp, err := service.UpdateBug(ctx, bugID, updateReq)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, repository.ErrBugNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Update error
	t.Run("Update Error", func(t *testing.T) {
		bugID := uint(1)
		existingBug := &model.Bug{
			ID:          bugID,
			Title:       "Existing Bug",
			Description: "Original description",
			Status:      model.BugStatusOpen,
			Priority:    model.BugPriorityMedium,
		}

		updateReq := &model.UpdateBugRequest{
			Title: strPtr("Updated Bug Title"),
		}

		mockRepo.On("GetByID", ctx, bugID).Return(existingBug, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Bug")).Return(errors.New("update error")).Once()

		resp, err := service.UpdateBug(ctx, bugID, updateReq)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "update error")
		mockRepo.AssertExpectations(t)
	})
}

// TestBugService_DeleteBug tests the DeleteBug method
func TestBugService_DeleteBug(t *testing.T) {
	service, mockRepo := setupBugServiceTest(t)
	ctx := context.Background()

	// Test case: Successful deletion
	t.Run("Successful Deletion", func(t *testing.T) {
		bugID := uint(1)
		mockRepo.On("Delete", ctx, bugID).Return(nil).Once()

		err := service.DeleteBug(ctx, bugID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Bug not found
	t.Run("Bug Not Found", func(t *testing.T) {
		bugID := uint(99)
		mockRepo.On("Delete", ctx, bugID).Return(repository.ErrBugNotFound).Once()

		err := service.DeleteBug(ctx, bugID)

		assert.Error(t, err)
		assert.Equal(t, repository.ErrBugNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

// TestBugService_ListBugs tests the ListBugs method
func TestBugService_ListBugs(t *testing.T) {
	service, mockRepo := setupBugServiceTest(t)
	ctx := context.Background()

	// Test case: Successful listing
	t.Run("Successful Listing", func(t *testing.T) {
		params := &model.BugListParams{
			Page:     1,
			PageSize: 10,
			Status:   model.BugStatusOpen,
		}

		mockBugs := []*model.Bug{
			{ID: 1, Title: "Bug 1", Status: model.BugStatusOpen},
			{ID: 2, Title: "Bug 2", Status: model.BugStatusOpen},
		}
		totalCount := int64(2)

		mockRepo.On("List", ctx, params).Return(mockBugs, totalCount, nil).Once()

		bugs, count, err := service.ListBugs(ctx, params)

		assert.NoError(t, err)
		assert.Equal(t, totalCount, count)
		assert.Len(t, bugs, 2)
		assert.Equal(t, uint(1), bugs[0].ID)
		assert.Equal(t, uint(2), bugs[1].ID)
		mockRepo.AssertExpectations(t)
	})

	// Test case: Repository error
	t.Run("Repository Error", func(t *testing.T) {
		params := &model.BugListParams{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, params).Return([]*model.Bug{}, int64(0), errors.New("list error")).Once()

		bugs, count, err := service.ListBugs(ctx, params)

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.Empty(t, bugs)
		assert.Contains(t, err.Error(), "list error")
		mockRepo.AssertExpectations(t)
	})
}

// Helper functions for pointer types
func strPtr(s string) *string {
	return &s
}

func statusPtr(s model.BugStatusType) *model.BugStatusType {
	return &s
}

func priorityPtr(p model.BugPriorityType) *model.BugPriorityType {
	return &p
}
