package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"EffiPlat/backend/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// pointyString returns a pointer to the string value s.
func pointyString(s string) *string {
	return &s
}

// pointyInt returns a pointer to the int value i.
func pointyInt(i int) *int {
	return &i
}

// Helper function to initialize a mock DB and repository for tests
func newTestServiceInstanceRepository(t *testing.T) (ServiceInstanceRepository, sqlmock.Sqlmock, *gorm.DB) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Expect the initial version query GORM makes for SQLite
	mock.ExpectQuery(regexp.QuoteMeta("select sqlite_version()")).WillReturnRows(sqlmock.NewRows([]string{"sqlite_version()"}).AddRow("3.30.1"))

	// Use a mock logger for tests to avoid noisy output, or a real one if logs are needed for debugging
	// testLogger := zap.NewNop()
	// For debugging, a real logger can be useful:
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // Or WarnLevel for less noise
	testLogger, _ := config.Build()

	// Create a GORM DB instance using the mock SQL connection.
	// Here we use SQLite dialector as an example, but it will be driven by sqlmock.
	// The actual database type doesn't matter much since sqlmock intercepts calls.
	gormDB, err := gorm.Open(sqlite.Dialector{DSN: "file::memory:?cache=shared", Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // GORM logger, not Zap
	})
	require.NoError(t, err)

	repo := NewServiceInstanceRepository(gormDB, testLogger)
	return repo, mock, gormDB // Return gormDB too if needed for direct interaction in tests
}

// This helper function is no longer needed since we've updated our test cases

func TestServiceInstanceRepositoryImpl_Create(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		now := time.Now()
		instance := &model.ServiceInstance{
			ServiceID:     1,
			EnvironmentID: 1,
			Version:       "1.0.0",
			Status:        model.ServiceInstanceStatusDeploying,
			DeployedAt:    &now,
			Config:        datatypes.JSONMap{}, // Initialize Config as an empty map
			// CreatedAt and UpdatedAt are usually set by GORM or DB
		}

		mock.ExpectBegin()
		insertQuery := "INSERT INTO `service_instances` (`service_id`,`environment_id`,`version`,`status`,`hostname`,`port`,`config`,`deployed_at`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(
				instance.ServiceID,
				instance.EnvironmentID,
				instance.Version,
				instance.Status,
				instance.Hostname, // This will be nil
				instance.Port,     // This will be nil
				instance.Config,   // Now an empty map, matching the error log
				instance.DeployedAt,
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
				nil,              // DeletedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1)) // LastInsertID, RowsAffected
		mock.ExpectCommit()

		err := repo.Create(context.Background(), instance)
		assert.NoError(t, err)
		assert.NotZero(t, instance.ID)                // GORM should set the ID after creation
		assert.NoError(t, mock.ExpectationsWereMet()) // Verify all expectations were met
	})

	t.Run("Database error on create", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		instance := &model.ServiceInstance{ServiceID: 1, EnvironmentID: 1, Version: "1.0.0"}

		mock.ExpectBegin()
		mock.ExpectExec(".*").WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback() // or ExpectCommit() if the transaction is not rolled back on error by GORM's Create

		err := repo.Create(context.Background(), instance)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrInvalidDB) // Check for specific error if possible
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceInstanceRepositoryImpl_GetByID(t *testing.T) {
	t.Run("Successful get by ID", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		now := time.Now()
		expectedInstance := &model.ServiceInstance{
			ID:            1,
			ServiceID:     10,
			EnvironmentID: 20,
			Version:       "1.0.1",
			Status:        model.ServiceInstanceStatusRunning,
			Hostname:      pointyString("prod-instance-1"),
			Port:          pointyInt(8080),
			DeployedAt:    &now,
			CreatedAt:     now.Add(-time.Hour),
			UpdatedAt:     now,
		}

		rows := sqlmock.NewRows([]string{"id", "service_id", "environment_id", "version", "status", "hostname", "port", "config", "deployed_at", "created_at", "updated_at", "deleted_at"}).
			AddRow(expectedInstance.ID, expectedInstance.ServiceID, expectedInstance.EnvironmentID, expectedInstance.Version, expectedInstance.Status, expectedInstance.Hostname, expectedInstance.Port, nil, expectedInstance.DeployedAt, expectedInstance.CreatedAt, expectedInstance.UpdatedAt, nil)

		// GORM's First method will use a query like: SELECT * FROM "service_instances" WHERE "service_instances"."id" = ? AND "service_instances"."deleted_at" IS NULL ORDER BY "service_instances"."id" LIMIT 1
		// We need to be careful with regexp.QuoteMeta if the query is complex or involves backticks that GORM adds.
		// For simplicity, we can match a broader pattern or ensure the exact GORM generated query is known.
		sqlQuery := "SELECT * FROM `service_instances` WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL ORDER BY `service_instances`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlQuery)).
			WithArgs(expectedInstance.ID).
			WillReturnRows(rows)

		instance, err := repo.GetByID(context.Background(), expectedInstance.ID)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, expectedInstance.ID, instance.ID)
		assert.Equal(t, expectedInstance.ServiceID, instance.ServiceID)
		assert.Equal(t, expectedInstance.Version, instance.Version)
		assert.Equal(t, expectedInstance.Status, instance.Status)
		assert.Equal(t, expectedInstance.Hostname, instance.Hostname)
		assert.Equal(t, expectedInstance.Port, instance.Port)
		// Timestamps can be tricky due to potential time zone differences or precision. Compare them carefully.
		assert.WithinDuration(t, *expectedInstance.DeployedAt, *instance.DeployedAt, time.Second)
		assert.WithinDuration(t, expectedInstance.CreatedAt, instance.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedInstance.UpdatedAt, instance.UpdatedAt, time.Second)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Get by ID - Not Found", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		id := uint(99)

		sqlQuery := "SELECT * FROM `service_instances` WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL ORDER BY `service_instances`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlQuery)).
			WithArgs(id).
			WillReturnError(gorm.ErrRecordNotFound)

		instance, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Get by ID - Database error", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		id := uint(1)

		sqlQuery := "SELECT * FROM `service_instances` WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL ORDER BY `service_instances`.`id` LIMIT 1"
		mock.ExpectQuery(regexp.QuoteMeta(sqlQuery)).
			WithArgs(id).
			WillReturnError(gorm.ErrInvalidDB) // Simulate a generic DB error

		instance, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceInstanceRepositoryImpl_List(t *testing.T) {
	t.Run("Successful list with pagination and default sorting", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		now := time.Now()

		expectedInstances := []model.ServiceInstance{
			{ID: 1, ServiceID: 1, EnvironmentID: 1, Version: "1.0.0", Status: model.ServiceInstanceStatusRunning, CreatedAt: now.Add(-time.Hour)},
			{ID: 2, ServiceID: 2, EnvironmentID: 1, Version: "1.0.1", Status: model.ServiceInstanceStatusDeploying, CreatedAt: now},
		}

		params := &ListServiceInstancesParams{
			Page:     1,
			PageSize: 10,
			// SortBy and Order will use defaults: createdAt, desc
		}

		// Mock Count Query
		countQuery := "SELECT count(*) FROM `service_instances` WHERE `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuery)).WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(len(expectedInstances)))

		// Mock Find Query (default sort: createdAt desc)
		// Note: The actual query GORM generates might have "service_instances"."deleted_at" IS NULL if soft delete is enabled on the model
		// GORM omits OFFSET 0 for the first page.
		findQuery := "SELECT * FROM `service_instances` WHERE `service_instances`.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT 10"
		rows := sqlmock.NewRows([]string{"id", "service_id", "environment_id", "version", "status", "created_at"})
		// Rows should be added in the order they are expected to be returned by the query (sorted)
		rows.AddRow(expectedInstances[1].ID, expectedInstances[1].ServiceID, expectedInstances[1].EnvironmentID, expectedInstances[1].Version, expectedInstances[1].Status, expectedInstances[1].CreatedAt) // newest first
		rows.AddRow(expectedInstances[0].ID, expectedInstances[0].ServiceID, expectedInstances[0].EnvironmentID, expectedInstances[0].Version, expectedInstances[0].Status, expectedInstances[0].CreatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(findQuery)).WillReturnRows(rows)

		instances, total, err := repo.List(context.Background(), params)

		assert.NoError(t, err)
		assert.Equal(t, int64(len(expectedInstances)), total)
		assert.Len(t, instances, len(expectedInstances))
		// Verify the order if necessary based on default sort (createdAt desc)
		assert.Equal(t, expectedInstances[1].ID, instances[0].ID)
		assert.Equal(t, expectedInstances[0].ID, instances[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Successful list with filters and custom sorting", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		now := time.Now()
		filterServiceID := uint(5)
		filterStatus := string(model.ServiceInstanceStatusRunning)
		sortByVersion := "version"
		orderAsc := "asc"

		expectedInstance := model.ServiceInstance{ID: 3, ServiceID: filterServiceID, EnvironmentID: 2, Version: "0.9.0", Status: model.ServiceInstanceStatusType(filterStatus), CreatedAt: now}

		params := &ListServiceInstancesParams{
			Page:      1,
			PageSize:  5,
			ServiceID: &filterServiceID,
			Status:    &filterStatus,
			SortBy:    sortByVersion,
			Order:     orderAsc,
		}

		// Mock Count Query with filters
		countQuery := "SELECT count(*) FROM `service_instances` WHERE service_id = ? AND status = ? AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuery)).
			WithArgs(filterServiceID, filterStatus).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))

		// Mock Find Query with filters and sorting
		// GORM omits OFFSET 0 for the first page.
		findQuery := "SELECT * FROM `service_instances` WHERE service_id = ? AND status = ? AND `service_instances`.`deleted_at` IS NULL ORDER BY version ASC LIMIT 5"
		rows := sqlmock.NewRows([]string{"id", "service_id", "environment_id", "version", "status", "created_at"}).
			AddRow(expectedInstance.ID, expectedInstance.ServiceID, expectedInstance.EnvironmentID, expectedInstance.Version, expectedInstance.Status, expectedInstance.CreatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(findQuery)).
			WithArgs(filterServiceID, filterStatus).
			WillReturnRows(rows)

		instances, total, err := repo.List(context.Background(), params)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, instances, 1)
		if len(instances) == 1 {
			assert.Equal(t, expectedInstance.ID, instances[0].ID)
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// TODO: Add test case for invalid status filter (should be ignored or return error based on implementation)
	// TODO: Add test case for empty result set
	// TODO: Add test case for database error on Count
	// TODO: Add test case for database error on Find
}

func TestServiceInstanceRepositoryImpl_Update(t *testing.T) {
	t.Run("Successful update", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		// Create a variable for the time first, then use its address
		now := time.Now()
		hostname := "updated-host"
		port := 8081
		
		instanceToUpdate := &model.ServiceInstance{
			ID:            1,
			ServiceID:     1,
			EnvironmentID: 1,
			Version:       "1.0.1",
			Status:        model.ServiceInstanceStatusRunning,
			Hostname:      &hostname,
			Port:          &port,
			Config:        map[string]interface{}{"key": "value"},
			DeployedAt:    &now,
		}

		// No transaction behavior with SkipDefaultTransaction
		mock.ExpectExec("UPDATE `service_instances` SET").
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

		err := repo.Update(context.Background(), instanceToUpdate)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update - Not Found because ID does not exist", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		instanceToUpdate := &model.ServiceInstance{ID: 99, Version: "1.0.0"} // Non-existent ID

		// No transaction behavior with SkipDefaultTransaction
		mock.ExpectExec("UPDATE `service_instances` SET").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected means record not found

		err := repo.Update(context.Background(), instanceToUpdate)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update - No fields changed (RowsAffected is 0, but record exists)", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		existingInstanceID := uint(1)
		instanceToUpdate := &model.ServiceInstance{ID: existingInstanceID, Version: "1.0.0"}

		// No transaction behavior with SkipDefaultTransaction
		mock.ExpectExec("UPDATE `service_instances` SET").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected means no changes

		// Based on the implementation, even when record exists but no changes, it returns ErrRecordNotFound
		err := repo.Update(context.Background(), instanceToUpdate)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update - Database error", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		instanceToUpdate := &model.ServiceInstance{ID: 1, Version: "1.0.1"}

		// No transaction behavior with SkipDefaultTransaction
		mock.ExpectExec("UPDATE `service_instances` SET").
			WillReturnError(gorm.ErrInvalidDB)

		err := repo.Update(context.Background(), instanceToUpdate)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update - instance ID is required", func(t *testing.T) {
		repo, _, _ := newTestServiceInstanceRepository(t)
		instanceToUpdate := &model.ServiceInstance{Version: "1.0.1"} // ID is 0

		err := repo.Update(context.Background(), instanceToUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instance ID is required")
		// No mock expectations as this is a pre-DB check
	})
}

func TestServiceInstanceRepositoryImpl_Delete(t *testing.T) {
	t.Run("Successful delete", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		idToDelete := uint(1)

		mock.ExpectBegin()
		// GORM's Delete method for soft delete typically updates `deleted_at`.
		// If it's a hard delete, it would be a DELETE FROM query.
		// Assuming soft delete given `gorm.DeletedAt` in the model.
		// The exact query can vary based on GORM version and configuration.
		// "UPDATE `service_instances` SET `deleted_at`=? WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL"
		deleteQuery := "UPDATE `service_instances` SET `deleted_at`=? WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).
			WithArgs(sqlmock.AnyArg(), idToDelete). // AnyArg for the timestamp
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), idToDelete)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete - Not Found", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		idToDelete := uint(99)

		mock.ExpectBegin()
		deleteQuery := "UPDATE `service_instances` SET `deleted_at`=? WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).
			WithArgs(sqlmock.AnyArg(), idToDelete).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		mock.ExpectCommit() // Changed from ExpectRollback() as GORM likely commits if no DB error occurs, even if 0 rows affected

		err := repo.Delete(context.Background(), idToDelete)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete - Database error", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		idToDelete := uint(1)

		mock.ExpectBegin()
		deleteQuery := "UPDATE `service_instances` SET `deleted_at`=? WHERE `service_instances`.`id` = ? AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).
			WithArgs(sqlmock.AnyArg(), idToDelete).
			WillReturnError(gorm.ErrInvalidDB)
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), idToDelete)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceInstanceRepositoryImpl_CheckExists(t *testing.T) {
	t.Run("CheckExists - Instance exists", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		serviceID := uint(1)
		environmentID := uint(1)
		version := "1.0.0"
		excludeID := uint(0) // Not excluding any ID

		countQuerySQL := "SELECT count(*) FROM `service_instances` WHERE (service_id = ? AND environment_id = ? AND version = ?) AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuerySQL)).
			WithArgs(serviceID, environmentID, version).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))

		exists, err := repo.CheckExists(context.Background(), serviceID, environmentID, version, excludeID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CheckExists - Instance does not exist", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		serviceID := uint(2)
		environmentID := uint(2)
		version := "2.0.0"
		excludeID := uint(0)

		countQuerySQL := "SELECT count(*) FROM `service_instances` WHERE (service_id = ? AND environment_id = ? AND version = ?) AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuerySQL)).
			WithArgs(serviceID, environmentID, version).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))

		exists, err := repo.CheckExists(context.Background(), serviceID, environmentID, version, excludeID)
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CheckExists - Instance exists but is excluded", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		serviceID := uint(3)
		environmentID := uint(3)
		version := "3.0.0"
		excludeID := uint(5) // Assuming instance with ID 5 is the one being checked against

		countQuerySQL := "SELECT count(*) FROM `service_instances` WHERE (service_id = ? AND environment_id = ? AND version = ?) AND id <> ? AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuerySQL)).
			WithArgs(serviceID, environmentID, version, excludeID).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0)) // Excluded, so count is 0 for *other* matches

		exists, err := repo.CheckExists(context.Background(), serviceID, environmentID, version, excludeID)
		assert.NoError(t, err)
		assert.False(t, exists) // Should be false as the match (if any) was excluded
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CheckExists - Database error", func(t *testing.T) {
		repo, mock, _ := newTestServiceInstanceRepository(t)
		serviceID := uint(4)
		environmentID := uint(4)
		version := "4.0.0"
		excludeID := uint(0)

		countQuerySQL := "SELECT count(*) FROM `service_instances` WHERE (service_id = ? AND environment_id = ? AND version = ?) AND `service_instances`.`deleted_at` IS NULL"
		mock.ExpectQuery(regexp.QuoteMeta(countQuerySQL)).
			WithArgs(serviceID, environmentID, version).
			WillReturnError(gorm.ErrInvalidDB)

		exists, err := repo.CheckExists(context.Background(), serviceID, environmentID, version, excludeID)
		assert.Error(t, err)
		assert.False(t, exists)
		assert.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
