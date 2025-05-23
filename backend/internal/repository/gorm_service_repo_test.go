package repository

import (
	"context"
	"EffiPlat/backend/internal/models"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Helper function to setup an in-memory SQLite database and repository for testing
func setupServiceTestDBAndRepo(t *testing.T) (*gorm.DB, ServiceRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate the schema
	err = gormDB.AutoMigrate(&models.ServiceType{}, &models.Service{})
	require.NoError(t, err)

	repo := NewGormServiceRepository(gormDB)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
		db.Close()
	}

	return gormDB, repo, mock, cleanup
}

// Helper to create a ServiceType for testing Services
func createTestServiceType(t *testing.T, db *gorm.DB, name string) *models.ServiceType {
	st := &models.ServiceType{Name: name, Description: "Test Service Type"}
	require.NoError(t, db.Create(st).Error)
	return st
}

func TestGormServiceRepository_Create(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TestTypeForServiceCreate")

	service := &models.Service{
		Name:          "Test Service One",
		Description:   "Description for service one",
		Version:       "1.0.0",
		Status:        models.ServiceStatusActive,
		ServiceTypeID: st1.ID,
	}

	err := repo.Create(ctx, service)
	assert.NoError(t, err)
	assert.NotZero(t, service.ID)

	// Test creating a service with the same name (should fail due to unique constraint if DB enforces it)
	// GORM SQLite might not enforce unique constraints defined in tags without explicit index creation
	// For this test, we assume the DB layer or GORM handles it. If not, service layer should catch it.
	// Here, we're testing the repository's Create method primarily.
}

func TestGormServiceRepository_GetByID(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TestTypeForServiceGetByID")

	createdService := &models.Service{
		Name:          "Test Service For GetByID",
		ServiceTypeID: st1.ID,
		Status:        models.ServiceStatusActive,
	}
	require.NoError(t, db.Create(createdService).Error)

	t.Run("found", func(t *testing.T) {
		retrievedService, err := repo.GetByID(ctx, createdService.ID)
		assert.NoError(t, err)
		require.NotNil(t, retrievedService)
		assert.Equal(t, createdService.Name, retrievedService.Name)
		assert.Equal(t, createdService.ServiceTypeID, retrievedService.ServiceTypeID)
		require.NotNil(t, retrievedService.ServiceType)
		assert.Equal(t, st1.Name, retrievedService.ServiceType.Name) // Check preloading
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrServiceNotFound))
	})
}

func TestGormServiceRepository_GetByName(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TestTypeForServiceGetByName")
	serviceName := "Unique Service Name For GetByName"

	createdService := &models.Service{
		Name:          serviceName,
		ServiceTypeID: st1.ID,
		Status:        models.ServiceStatusActive,
	}
	require.NoError(t, db.Create(createdService).Error)

	t.Run("found", func(t *testing.T) {
		retrievedService, err := repo.GetByName(ctx, serviceName)
		assert.NoError(t, err)
		require.NotNil(t, retrievedService)
		assert.Equal(t, serviceName, retrievedService.Name)
	})

	t.Run("not found", func(t *testing.T) {
		retrievedService, err := repo.GetByName(ctx, "NonExistentServiceName")
		assert.NoError(t, err) // GetByName returns nil, nil for not found as per its contract
		assert.Nil(t, retrievedService)
	})
}

func TestGormServiceRepository_List(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TypeA_List")
	st2 := createTestServiceType(t, db, "TypeB_List")

	// Create some services
	servicesData := []models.Service{
		{Name: "Service Alpha", ServiceTypeID: st1.ID, Status: models.ServiceStatusActive},
		{Name: "Service Beta", ServiceTypeID: st1.ID, Status: models.ServiceStatusInactive},
		{Name: "Service Gamma", ServiceTypeID: st2.ID, Status: models.ServiceStatusActive},
		{Name: "Another Alpha", ServiceTypeID: st2.ID, Status: models.ServiceStatusDevelopment},
	}

	for _, s := range servicesData {
		// Need to capture s by value for the pointer
		svc := s
		require.NoError(t, db.Create(&svc).Error)
		time.Sleep(10 * time.Millisecond) // Ensure CreatedAt is different for ordering tests
	}

	t.Run("list all no filters", func(t *testing.T) {
		params := models.ServiceListParams{Page: 1, PageSize: 10}
		services, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(servicesData)), total)
		assert.Len(t, services, len(servicesData))
		// Check if ServiceType is preloaded
		for _, s := range services {
			assert.NotNil(t, s.ServiceType, "ServiceType should be preloaded for service %s", s.Name)
		}
	})

	t.Run("list with name filter", func(t *testing.T) {
		params := models.ServiceListParams{Page: 1, PageSize: 10, Name: "Alpha"}
		services, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, services, 2)
		for _, s := range services {
			assert.Contains(t, s.Name, "Alpha")
		}
	})

	t.Run("list with status filter", func(t *testing.T) {
		params := models.ServiceListParams{Page: 1, PageSize: 10, Status: string(models.ServiceStatusActive)}
		services, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, services, 2)
		for _, s := range services {
			assert.Equal(t, models.ServiceStatusActive, s.Status)
		}
	})

	t.Run("list with service type id filter", func(t *testing.T) {
		params := models.ServiceListParams{Page: 1, PageSize: 10, ServiceTypeID: st1.ID}
		services, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, services, 2)
		for _, s := range services {
			assert.Equal(t, st1.ID, s.ServiceTypeID)
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		paramsPage1 := models.ServiceListParams{Page: 1, PageSize: 2}
		services1, total1, err1 := repo.List(ctx, paramsPage1)
		assert.NoError(t, err1)
		assert.Equal(t, int64(len(servicesData)), total1)
		assert.Len(t, services1, 2)

		paramsPage2 := models.ServiceListParams{Page: 2, PageSize: 2}
		services2, total2, err2 := repo.List(ctx, paramsPage2)
		assert.NoError(t, err2)
		assert.Equal(t, int64(len(servicesData)), total2)
		assert.Len(t, services2, 2)

		// Ensure no overlap between pages (assuming default order by CreatedAt DESC)
		assert.NotEqual(t, services1[0].ID, services2[0].ID)
		assert.NotEqual(t, services1[1].ID, services2[1].ID)
	})

	t.Run("list empty result", func(t *testing.T) {
		params := models.ServiceListParams{Page: 1, PageSize: 10, Name: "NonExistentNameForList"}
		services, total, err := repo.List(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, services, 0)
	})
}

func TestGormServiceRepository_Update(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TestTypeForServiceUpdate1")
	st2 := createTestServiceType(t, db, "TestTypeForServiceUpdate2")

	createdService := &models.Service{
		Name:          "Service Before Update",
		Description:   "Initial Description",
		Version:       "1.0",
		Status:        models.ServiceStatusDevelopment,
		ServiceTypeID: st1.ID,
	}
	require.NoError(t, db.Create(createdService).Error)

	updatedName := "Service After Update"
	updatedDescription := "Updated Description"
	updatedVersion := "2.0"
	updatedStatus := models.ServiceStatusActive

	// Make a copy to update
	serviceToUpdate := *createdService
	serviceToUpdate.Name = updatedName
	serviceToUpdate.Description = updatedDescription
	serviceToUpdate.Version = updatedVersion
	serviceToUpdate.Status = updatedStatus
	serviceToUpdate.ServiceTypeID = st2.ID // Change service type

	err := repo.Update(ctx, &serviceToUpdate)
	assert.NoError(t, err)

	retrievedService, _ := repo.GetByID(ctx, createdService.ID)
	require.NotNil(t, retrievedService)
	assert.Equal(t, updatedName, retrievedService.Name)
	assert.Equal(t, updatedDescription, retrievedService.Description)
	assert.Equal(t, updatedVersion, retrievedService.Version)
	assert.Equal(t, updatedStatus, retrievedService.Status)
	assert.Equal(t, st2.ID, retrievedService.ServiceTypeID)
	require.NotNil(t, retrievedService.ServiceType)
	assert.Equal(t, st2.Name, retrievedService.ServiceType.Name)
}

func TestGormServiceRepository_Delete(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "TestTypeForServiceDelete")

	createdService := &models.Service{
		Name:          "Service To Be Deleted",
		ServiceTypeID: st1.ID,
		Status:        models.ServiceStatusActive,
	}
	require.NoError(t, db.Create(createdService).Error)

	err := repo.Delete(ctx, createdService.ID)
	assert.NoError(t, err)

	// Verify it's soft-deleted (cannot be retrieved by GetByID normally)
	_, err = repo.GetByID(ctx, createdService.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrServiceNotFound))

	// Verify with Unscoped to find soft-deleted record
	var softDeletedService models.Service
	err = db.Unscoped().First(&softDeletedService, createdService.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, softDeletedService.DeletedAt)

	// Test deleting non-existent service
	err = repo.Delete(ctx, 99999)
	assert.NoError(t, err) // GORM's Delete doesn't error if record not found by default
}

func TestGormServiceRepository_CountServicesByServiceTypeID(t *testing.T) {
	db, repo, _, cleanup := setupServiceTestDBAndRepo(t)
	defer cleanup()

	ctx := context.Background()
	st1 := createTestServiceType(t, db, "Type1_Count")
	st2 := createTestServiceType(t, db, "Type2_Count")
	st3 := createTestServiceType(t, db, "Type3_NoServices_Count")

	// Services for st1
	_ = db.Create(&models.Service{Name: "S1T1", ServiceTypeID: st1.ID, Status: models.ServiceStatusActive})
	_ = db.Create(&models.Service{Name: "S2T1", ServiceTypeID: st1.ID, Status: models.ServiceStatusActive})

	// Service for st2
	_ = db.Create(&models.Service{Name: "S1T2", ServiceTypeID: st2.ID, Status: models.ServiceStatusActive})

	t.Run("count for type with services", func(t *testing.T) {
		count, err := repo.CountServicesByServiceTypeID(ctx, st1.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = repo.CountServicesByServiceTypeID(ctx, st2.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("count for type with no services", func(t *testing.T) {
		count, err := repo.CountServicesByServiceTypeID(ctx, st3.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("count for non-existent service type id", func(t *testing.T) {
		count, err := repo.CountServicesByServiceTypeID(ctx, 99999)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// MockLogger for GORM to suppress logs during tests or capture them
type MockLogger struct{}

func (m *MockLogger) LogMode(level logger.LogLevel) logger.Interface { return m }
func (m *MockLogger) Info(ctx context.Context, s string, args ...interface{})  {}
func (m *MockLogger) Warn(ctx context.Context, s string, args ...interface{})  {}
func (m *MockLogger) Error(ctx context.Context, s string, args ...interface{}) {}
func (m *MockLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
}

// Helper function to escape strings for SQL LIKE queries if using sqlmock directly
func escapeSQLLike(s string) string {
	return regexp.QuoteMeta(s)
}
