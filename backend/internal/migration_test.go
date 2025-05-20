package internal_test

import (
	"database/sql"
	"embed"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite Driver - underscore import for side effects

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// setupTestDB creates an in-memory SQLite DB and a migrate instance for testing.
func setupTestDB(t *testing.T) (*sql.DB, *migrate.Migrate) {
	// Use in-memory SQLite for isolated testing
	// Removed _x_no_tx=1 from DSN as we'll try NoTxWrap in sqlite3.Config
	dsn := "file::memory:?cache=shared"
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err, "Failed to open in-memory db")

	// Setup migrate instance
	sourceDriver, err := iofs.New(migrationsFS, "migrations") // Load from embedded FS, root is migrations
	require.NoError(t, err, "Failed to create source driver")

	// Use NoTxWrap: true in sqlite3.Config to prevent migrate from wrapping migrations in transactions.
	dbDriver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{NoTxWrap: true})
	require.NoError(t, err, "Failed to create database driver")

	migrateInstance, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	require.NoError(t, err, "Failed to create migrate instance")

	// Cleanup function to close DB connection after test
	t.Cleanup(func() {
		srcErr, dbErr := migrateInstance.Close() // Close source and database drivers
		require.NoError(t, srcErr, "Failed to close source driver")
		require.NoError(t, dbErr, "Failed to close database driver")

		err := sqlDB.Close()
		require.NoError(t, err, "Failed to close db connection")
	})

	return sqlDB, migrateInstance
}

// TestInitialMigration tests the first migration script (up and down).
func TestInitialMigration(t *testing.T) {
	db, m := setupTestDB(t)

	// 1. Test Up migration
	t.Run("ApplyUpMigration", func(t *testing.T) {
		err := m.Up()
		require.NoError(t, err, "Failed to apply Up migration")

		// Verify schema: Check if 'users' table exists
		_, err = db.Exec("SELECT id FROM users LIMIT 1") // Simple check by trying to query
		require.NoError(t, err, "Table 'users' should exist after migration")

		// Verify schema_migrations table
		var version uint
		var dirty bool
		err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
		require.NoError(t, err, "Failed to query schema_migrations")
		require.Equal(t, uint(1), version, "Schema version should be 1")
		require.False(t, dirty, "Schema should not be dirty")

		// Add more checks: table existence, column existence, index existence etc.
		// Example using PRAGMA:
		rows, err := db.Query("PRAGMA table_info('users')")
		require.NoError(t, err)
		defer rows.Close()
		columnCount := 0
		foundNameCol := false
		for rows.Next() {
			columnCount++
			var cid int
			var name string
			var typeName string
			var notnull int
			var dfltValue sql.NullString
			var pk int
			err = rows.Scan(&cid, &name, &typeName, &notnull, &dfltValue, &pk)
			require.NoError(t, err)
			if name == "name" {
				foundNameCol = true
			}
		}
		require.Greater(t, columnCount, 5, "Users table should have several columns")
		require.True(t, foundNameCol, "Users table should have a 'name' column")
	})

	// 2. Test Down migration
	t.Run("ApplyDownMigration", func(t *testing.T) {
		// We assume the Up migration ran successfully in the previous subtest
		// or we could run m.Up() here again if subtests need to be independent.

		err := m.Down() // Rollback the last migration (version 1)
		require.NoError(t, err, "Failed to apply Down migration")

		// Verify schema: Check if 'users' table no longer exists
		_, err = db.Exec("SELECT id FROM users LIMIT 1")
		require.Error(t, err, "Table 'users' should not exist after rollback") // Expect an error
		require.Contains(t, err.Error(), "no such table: users", "Error should indicate table not found")

		// Verify schema_migrations table is empty or gone
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1 AND dirty = 0").Scan(&count)
		// We might get an error if the table itself is dropped by `down -all`, adjust check if needed
		if err == nil {
			require.Equal(t, 0, count, "Schema version 1 should be marked as not applied")
		} else {
			require.Contains(t, err.Error(), "no such table: schema_migrations", "schema_migrations might be dropped or inaccessible")
		}
	})

	// 3. Test reapplying Up migration
	t.Run("ReapplyUpMigration", func(t *testing.T) {
		// We assume Down migration ran successfully

		err := m.Up()
		require.NoError(t, err, "Failed to reapply Up migration")

		// Verify schema again (similar to step 1)
		_, err = db.Exec("SELECT id FROM users LIMIT 1")
		require.NoError(t, err, "Table 'users' should exist after reapplying migration")

		// Verify schema_migrations again
		var version uint
		var dirty bool
		err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
		require.NoError(t, err, "Failed to query schema_migrations after reapply")
		require.Equal(t, uint(1), version, "Schema version should be 1 after reapply")
		require.False(t, dirty, "Schema should not be dirty after reapply")
	})
}
