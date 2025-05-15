package seed

import (
	"EffiPlat/backend/internal/factories"
	"EffiPlat/backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// SeedRoles creates some sample roles.
func SeedRoles(db *gorm.DB) error {
	fmt.Println("Seeding roles...")

	roles := []models.Role{
		{Name: "admin", Description: "System administrator with full access"},
		{Name: "user", Description: "Regular user with limited access"},
		{Name: "manager", Description: "Team manager with elevated access"},
	}

	for _, role := range roles {
		if err := db.FirstOrCreate(&role, models.Role{Name: role.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed role %s: %w", role.Name, err)
		}
	}

	fmt.Println("Role seeding complete.")
	return nil
}

// SeedUsers creates some sample users.
func SeedUsers(db *gorm.DB) error {
	fmt.Println("Seeding users...")

	// First, ensure roles exist
	if err := SeedRoles(db); err != nil {
		return err
	}

	// Get the admin role
	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("failed to find admin role: %w", err)
	}

	// Create an admin user
	_, err := factories.NewUserFactory().
		WithName("Admin User").
		WithEmail("admin@effiplat.local").
		WithPassword("password"). // Use a strong password in real scenarios
		WithRoles([]models.Role{adminRole}).
		Create(db)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Get the regular user role
	var userRole models.Role
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		return fmt.Errorf("failed to find user role: %w", err)
	}

	// Create a few standard users
	for i := 0; i < 5; i++ {
		_, err := factories.NewUserFactory().
			WithName(fmt.Sprintf("Test User %d", i+1)).
			WithEmail(fmt.Sprintf("testuser%d@example.com", i+1)).
			WithRoles([]models.Role{userRole}).
			Create(db)
		if err != nil {
			return fmt.Errorf("failed to seed test user %d: %w", i+1, err)
		}
	}

	fmt.Println("User seeding complete.")
	return nil
}

// SeedAll runs all the seeders.
func SeedAll(db *gorm.DB) error {
	fmt.Println("Starting database seeding...")

	if err := SeedUsers(db); err != nil {
		return err
	}

	// Add calls to other seeders here later, e.g.:
	// if err := SeedEnvironments(db); err != nil {
	//     return err
	// }
	// if err := SeedServices(db); err != nil {
	//     return err
	// }

	fmt.Println("Database seeding finished successfully.")
	return nil
}
