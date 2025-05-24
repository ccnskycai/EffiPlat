package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
	"errors"
	"regexp" // For sqlmock RegexpMatch
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/postgres" // Assuming PostgreSQL, adjust if different
	"gorm.io/gorm"
)

// Helper function to setup mock DB and repository for tests
func setupBugRepositoryTest(t *testing.T) (BugRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	// Create a mock logger for testing
	logger, _ := zap.NewDevelopment()

	repo := NewBugRepository(gormDB, logger)

	return repo, mock, func() {
		mockDB.Close()
		logger.Sync()
	}
}

// TestBugRepositoryImpl_Create tests the Create method of BugRepository.
func TestBugRepositoryImpl_Create(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()
	bugToCreate := &model.Bug{
		Title:       "Test Bug",
		Description: "A bug for testing",
		Status:      model.BugStatusOpen,
		Priority:    model.BugPriorityMedium,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "bugs" ("title","description","status","priority","reporter_id","assignee_id","created_at","updated_at","deleted_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
		WithArgs(bugToCreate.Title, bugToCreate.Description, bugToCreate.Status, bugToCreate.Priority, bugToCreate.ReporterID, bugToCreate.AssigneeID, bugToCreate.CreatedAt, bugToCreate.UpdatedAt, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(ctx, bugToCreate)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), bugToCreate.ID) // Check if ID is set
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_GetByID_Success tests successful retrieval by ID.
func TestBugRepositoryImpl_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(1)
	now := time.Now()
	expectedBug := &model.Bug{
		ID:          bugID,
		Title:       "Found Bug",
		Description: "This bug was found",
		Status:      model.BugStatusOpen,
		Priority:    model.BugPriorityHigh,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "priority", "reporter_id", "assignee_id", "created_at", "updated_at", "deleted_at"}).
		AddRow(expectedBug.ID, expectedBug.Title, expectedBug.Description, expectedBug.Status, expectedBug.Priority, nil, nil, expectedBug.CreatedAt, expectedBug.UpdatedAt, nil)

	// Explicitly handle both arguments (ID and LIMIT)
	mock.ExpectQuery(`SELECT \* FROM "bugs" WHERE "bugs"\."id" = \$1 AND "bugs"\."deleted_at" IS NULL ORDER BY "bugs"\."id" LIMIT \$2`).
		WithArgs(bugID, 1).
		WillReturnRows(rows)

	bug, err := repo.GetByID(ctx, bugID)
	assert.NoError(t, err)
	assert.NotNil(t, bug)
	assert.Equal(t, expectedBug.ID, bug.ID)
	assert.Equal(t, expectedBug.Title, bug.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_GetByID_NotFound tests retrieval of a non-existent bug.
func TestBugRepositoryImpl_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(99)

	// Explicitly handle both arguments (ID and LIMIT)
	mock.ExpectQuery(`SELECT \* FROM "bugs" WHERE "bugs"\."id" = \$1 AND "bugs"\."deleted_at" IS NULL ORDER BY "bugs"\."id" LIMIT \$2`).
		WithArgs(bugID, 1).
		WillReturnError(gorm.ErrRecordNotFound) // Simulate GORM's not found error

	bug, err := repo.GetByID(ctx, bugID)
	assert.Error(t, err)
	assert.Nil(t, bug)
	assert.True(t, errors.Is(err, ErrBugNotFound), "Expected ErrBugNotFound")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// AnyTime argument matcher for sqlmock for time.Time fields
// GORM often sets these on create/update so exact match is hard.
// type AnyTime struct{}

// // Match satisfies sqlmock.Argument interface
// func (a AnyTime) Match(v driver.Value) bool {
// 	_, ok := v.(time.Time)
// 	return ok
// }

// TestBugRepositoryImpl_Update tests the update method
func TestBugRepositoryImpl_Update(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(1)
	bugToUpdate := &model.Bug{
		ID:          bugID,
		Title:       "Updated Bug Title",
		Description: "Updated description",
		Status:      model.BugStatusInProgress,
		Priority:    model.BugPriorityHigh,
	}

	// Expect the query to check if the bug exists
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE \(id = \$1 AND deleted_at IS NULL\) AND "bugs"\."deleted_at" IS NULL`).
		WithArgs(bugID).
		WillReturnRows(countRows)

	mock.ExpectBegin()
	// Use a regex pattern to match the UPDATE query with greater flexibility
	mock.ExpectExec(`UPDATE "bugs" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, bugToUpdate)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_Update_NotFound tests update when bug doesn't exist
func TestBugRepositoryImpl_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(99) // Non-existent ID
	bugToUpdate := &model.Bug{
		ID:       bugID,
		Title:    "Updated Bug Title",
		Status:   model.BugStatusInProgress,
		Priority: model.BugPriorityHigh,
	}

	// Expect the query to check if the bug exists, return 0 to indicate not found
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE \(id = \$1 AND deleted_at IS NULL\) AND "bugs"\."deleted_at" IS NULL`).
		WithArgs(bugID).
		WillReturnRows(countRows)

	// No BEGIN or other SQL expected as the function should return early

	err := repo.Update(ctx, bugToUpdate)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrBugNotFound), "Expected ErrBugNotFound")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_Delete tests the delete method
func TestBugRepositoryImpl_Delete(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(1)

	// Expect the query to check if the bug exists
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE \(id = \$1 AND deleted_at IS NULL\) AND "bugs"\."deleted_at" IS NULL`).
		WithArgs(bugID).
		WillReturnRows(countRows)

	mock.ExpectBegin()
	// Use a regex pattern to match the soft delete UPDATE query
	mock.ExpectExec(`UPDATE "bugs" SET "deleted_at"=`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, bugID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_Delete_NotFound tests delete when bug doesn't exist
func TestBugRepositoryImpl_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	bugID := uint(99) // Non-existent ID

	// Expect the query to check if the bug exists, return 0 to indicate not found
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE \(id = \$1 AND deleted_at IS NULL\) AND "bugs"\."deleted_at" IS NULL`).
		WithArgs(bugID).
		WillReturnRows(countRows)

	// No BEGIN or other SQL expected as the function should return early

	err := repo.Delete(ctx, bugID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrBugNotFound), "Expected ErrBugNotFound")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_List tests listing bugs with filters and pagination
func TestBugRepositoryImpl_List(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	params := &model.BugListParams{
		Page:     1,
		PageSize: 10,
		Title:    "Bug",
		Status:   model.BugStatusOpen,
	}

	// Mock for count query with more flexible regex pattern
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE title LIKE`).
		WillReturnRows(countRows)

	// Mock for list query with pagination, using a more flexible pattern
	listRows := sqlmock.NewRows([]string{"id", "title", "description", "status", "priority", "reporter_id", "assignee_id", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Bug 1", "Description 1", model.BugStatusOpen, model.BugPriorityHigh, nil, nil, time.Now(), time.Now(), nil).
		AddRow(2, "Bug 2", "Description 2", model.BugStatusOpen, model.BugPriorityMedium, nil, nil, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`SELECT \* FROM "bugs" WHERE title LIKE`).
		WillReturnRows(listRows)

	bugs, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, bugs, 2)
	assert.Equal(t, "Bug 1", bugs[0].Title)
	assert.Equal(t, "Bug 2", bugs[1].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_CountBugsByAssigneeID tests counting bugs by assignee ID
func TestBugRepositoryImpl_CountBugsByAssigneeID(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	assigneeID := uint(5)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE assignee_id = \$1`).
		WithArgs(assigneeID).
		WillReturnRows(countRows)

	count, err := repo.CountBugsByAssigneeID(ctx, assigneeID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_CountBugsByEnvironmentID tests counting bugs by environment ID
func TestBugRepositoryImpl_CountBugsByEnvironmentID(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	environmentID := uint(2)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE environment_id = \$1`).
		WithArgs(environmentID).
		WillReturnRows(countRows)

	count, err := repo.CountBugsByEnvironmentID(ctx, environmentID)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBugRepositoryImpl_GetBugsByStatus tests getting bugs by status
func TestBugRepositoryImpl_GetBugsByStatus(t *testing.T) {
	repo, mock, cleanup := setupBugRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	status := model.BugStatusInProgress
	params := &model.BugListParams{
		Page:     1,
		PageSize: 10,
	}

	// Mock for count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "bugs" WHERE status = \$1`).
		WillReturnRows(countRows)

	// Mock for list query with pagination
	listRows := sqlmock.NewRows([]string{"id", "title", "description", "status", "priority", "reporter_id", "assignee_id", "created_at", "updated_at", "deleted_at"}).
		AddRow(3, "Bug 3", "Description 3", status, model.BugPriorityHigh, nil, nil, time.Now(), time.Now(), nil).
		AddRow(4, "Bug 4", "Description 4", status, model.BugPriorityHigh, nil, nil, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`SELECT \* FROM "bugs" WHERE status = \$1`).
		WillReturnRows(listRows)

	bugs, total, err := repo.GetBugsByStatus(ctx, status, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, bugs, 2)
	assert.Equal(t, status, bugs[0].Status)
	assert.Equal(t, status, bugs[1].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}
