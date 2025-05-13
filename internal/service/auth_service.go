package service

import (
	"errors"
	"time"

	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthService provides authentication-related services.
type AuthService struct {
	DB *gorm.DB
	jwtKey []byte
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(db *gorm.DB, jwtKey []byte) *AuthService {
	return &AuthService{
		DB:     db,
		jwtKey: jwtKey,
	}
}

// Login authenticates a user and returns a JWT token upon successful authentication.
func (s *AuthService) Login(email, password string) (string, error) {
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// It's generally better to return a generic error message for invalid credentials
			// to avoid leaking information about whether an email exists in the system.
			return "", errors.New("invalid email or password")
		}
		// For other database errors, return the error directly or a generic server error.
		return "", err
	}

	// Compare the provided password with the stored password hash.
	// In a real application, passwords should always be hashed using a strong algorithm like bcrypt.
	// The following is a placeholder for plain text comparison and should be replaced.
	// err := utils.ComparePassword(user.Password, password) // Assuming user.Password stores the hash
	// if err != nil {
	// return "", errors.New("invalid email or password")
	// }

	// FIXME: Replace with actual password hash comparison
	// This is a temporary plain text comparison and is insecure.
	if user.Password != password { // THIS IS THE LINE TO BE MODIFIED IN THE FUTURE for hashing
		return "", errors.New("invalid email or password")
	}

	// Define JWT claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
		// Add other claims as needed, e.g., roles
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return t, nil
}

// Register creates a new user.
// FIXME: Implement user registration logic, including password hashing.
func (s *AuthService) Register(user *models.User) error {
	// Hash the password before saving to the database
	hashedPassword, err := utils.HashPassword(user.Password) // user.Password is plain text here
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Attempt to create the user record in the database.
	if err := s.DB.Create(user).Error; err != nil {
		// You might want to check for specific database errors here,
		// like a duplicate email violation, and return a more specific error message.
		return err
	}
	return nil
}
