//go:generate wire
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

// wire.go - now in internal package

// ProviderSet for user components
var UserSet = wire.NewSet(
	repository.NewUserRepository,
	service.NewUserService,
	handler.NewUserHandler,
)

// InitializeUserHandler is the injector for UserHandler and its dependencies.
// It takes the database connection as input.
// This function will be callable from other packages (like main) because it's exported.
func InitializeUserHandler(db *gorm.DB) (*handler.UserHandler, error) {
	wire.Build(
		UserSet,
	)
	return nil, nil // These will be replaced by Wire
}

// ProviderSet for auth components
var AuthSet = wire.NewSet(
	repository.NewUserRepository, // Shared, or could be in a common set
	service.NewAuthService,
	handler.NewAuthHandler,
	// Potentially add wire.Bind here if NewAuthService returns concrete but needs interface, etc.
)

// InitializeAuthHandler is the injector for AuthHandler.
// Make sure it has the //go:build wireinject tags if it's in a wireinject file.
// If wire.go is itself a wireinject file (based on build tags at the top),
// then this function template is fine.
func InitializeAuthHandler(db *gorm.DB, jwtKey []byte, logger *zap.Logger) (*handler.AuthHandler, error) {
	wire.Build(
		AuthSet,
		// If NewUserRepository needs logger, and logger is provided to InitializeAuthHandler,
		// Wire will connect them. Ensure all dependencies for NewUserRepository,
		// NewAuthService, NewAuthHandler are available as parameters to this
		// InitializeAuthHandler or are provided by other providers in the AuthSet.
		// For example, if NewAuthService needs jwtKey and logger:
		// wire.Value(jwtKey), // This makes jwtKey available if it's a simple value
		// wire.Value(logger),  // This makes logger available
	)
	return nil, nil // Wire will replace this
}
