//go:build wireinject
// +build wireinject

package internal

import (
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/service"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InitAuthHandler(db *gorm.DB, jwtKey []byte, logger *zap.Logger) *handler.AuthHandler {
	wire.Build(
		repository.NewUserRepository,
		service.NewAuthService,
		handler.NewAuthHandler,
	)
	return &handler.AuthHandler{}
}
