package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo repository.UserRepository
	jwtKey   []byte
	logger   *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtKey []byte, logger *zap.Logger) *AuthService {
	return &AuthService{userRepo: userRepo, jwtKey: jwtKey, logger: logger}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.LoginResponse, error) {
	s.logger.Info("Login attempt", zap.String("email", email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found by email", zap.String("email", email), zap.Error(err))
			return nil, utils.ErrInvalidCredentials
		}
		s.logger.Error("Error fetching user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	if user == nil {
		s.logger.Warn("User object is nil after FindByEmail (no error)", zap.String("email", email))
		return nil, utils.ErrInvalidCredentials
	}

	s.logger.Info("User found", zap.Uint("userID", user.ID), zap.String("email", user.Email))

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("Password comparison failed", zap.Uint("userID", user.ID), zap.String("email", email), zap.Error(err))
		return nil, utils.ErrInvalidCredentials
	}

	s.logger.Info("Password comparison successful", zap.Uint("userID", user.ID), zap.String("email", email))

	claims := model.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		s.logger.Error("Failed to sign JWT token", zap.Uint("userID", user.ID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Login successful, token generated", zap.Uint("userID", user.ID), zap.String("email", email))

	return &model.LoginResponse{
		Token: tokenString,
		User:  user,
	}, nil
}
