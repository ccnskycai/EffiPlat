package seed

import (
	"EffiPlat/backend/internal/factories"
	"fmt"

	"gorm.io/gorm"
)

// SeedUsers creates some sample users.
func SeedUsers(db *gorm.DB) error {
	fmt.Println("Seeding users...")

	// Create an admin user
	_, err := factories.NewUserFactory().
		WithName("Admin User").
		WithEmail("admin@effiplat.local").
		WithPassword("adminpassword"). // Use a strong password in real scenarios
		Create(db)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Create a few standard users
	for i := 0; i < 5; i++ {
		_, err := factories.NewUserFactory().
			WithName(fmt.Sprintf("Test User %d", i+1)).
			WithEmail(fmt.Sprintf("testuser%d@example.com", i+1)).
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
