//go:generate wire
//go:build wireinject
// +build wireinject

// To regenerate wire_gen.go, navigate to the `backend` directory and run:
// go generate ./...
// If you encounter issues, ensure wire CLI is installed and in your PATH, then you might try:
// cd backend && wire

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
	repository.NewRoleRepository,
	service.NewUserService,
	handler.NewUserHandler,
	wire.Bind(new(repository.RoleRepository), new(*repository.RoleRepositoryImpl)),
)

// InitializeUserHandler is the injector for UserHandler and its dependencies.
// It takes the database connection as input.
// This function will be callable from other packages (like main) because it's exported.
func InitializeUserHandler(db *gorm.DB, logger *zap.Logger) (*handler.UserHandler, error) {
	wire.Build(
		UserSet,
	)
	return nil, nil // These will be replaced by Wire
}

// ProviderSet for role components
var RoleSet = wire.NewSet(
	repository.NewRoleRepository,
	service.NewRoleService,
	handler.NewRoleHandler,
	wire.Bind(new(repository.RoleRepository), new(*repository.RoleRepositoryImpl)), // Updated to RoleRepositoryImpl
	wire.Bind(new(service.RoleService), new(*service.RoleServiceImpl)),             // Updated to RoleServiceImpl
)

// InitializeRoleHandler is the injector for RoleHandler and its dependencies.
func InitializeRoleHandler(db *gorm.DB, logger *zap.Logger) (*handler.RoleHandler, error) {
	wire.Build(
		RoleSet,
		// We provide db and logger as parameters to this injector, so they are available
		// to NewRoleRepository, NewRoleService, and NewRoleHandler if their constructors need them.
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for auth components
var AuthSet = wire.NewSet(
	repository.NewUserRepository, // This now returns the interface type
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

// ProviderSet for permission components
var PermissionSet = wire.NewSet(
	repository.NewPermissionRepository,
	wire.Bind(new(repository.PermissionRepository), new(*repository.PermissionRepositoryImpl)),

	repository.NewRoleRepository, // Needed by NewPermissionService
	wire.Bind(new(repository.RoleRepository), new(*repository.RoleRepositoryImpl)),

	service.NewPermissionService, // Provider returns the interface type, explicit bind for its own interface is redundant

	handler.NewPermissionHandler,
)

// InitializePermissionHandler is the injector for PermissionHandler and its dependencies.
func InitializePermissionHandler(db *gorm.DB, logger *zap.Logger) (*handler.PermissionHandler, error) {
	wire.Build(
		PermissionSet,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for responsibility components
var ResponsibilitySet = wire.NewSet(
	repository.NewGormResponsibilityRepository,
	service.NewResponsibilityService,
	handler.NewResponsibilityHandler,
)

// InitializeResponsibilityHandler is the injector for ResponsibilityHandler and its dependencies.
func InitializeResponsibilityHandler(db *gorm.DB, logger *zap.Logger) (*handler.ResponsibilityHandler, error) {
	wire.Build(
		ResponsibilitySet,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for responsibility group components
var ResponsibilityGroupSet = wire.NewSet(
	repository.NewGormResponsibilityGroupRepository,
	repository.NewGormResponsibilityRepository, // For validation in service
	service.NewResponsibilityGroupService,
	handler.NewResponsibilityGroupHandler,
)

// InitializeResponsibilityGroupHandler is the injector for ResponsibilityGroupHandler and its dependencies.
func InitializeResponsibilityGroupHandler(db *gorm.DB, logger *zap.Logger) (*handler.ResponsibilityGroupHandler, error) {
	wire.Build(
		ResponsibilityGroupSet,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for Environment components
var EnvironmentSet = wire.NewSet(
	repository.NewGormEnvironmentRepository,
	service.NewEnvironmentService,
	handler.NewEnvironmentHandler,
)

// InitializeEnvironmentHandler is the injector for EnvironmentHandler and its dependencies.
func InitializeEnvironmentHandler(db *gorm.DB, logger *zap.Logger) (*handler.EnvironmentHandler, error) {
	wire.Build(
		EnvironmentSet,
	)
	return nil, nil // Wire will replace this
}

// InitializeEnvironmentRepository is the injector for EnvironmentRepository.
func InitializeEnvironmentRepository(db *gorm.DB, logger *zap.Logger) (repository.EnvironmentRepository, error) {
	wire.Build(
		repository.NewGormEnvironmentRepository,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for Asset components
var AssetSet = wire.NewSet(
	repository.NewGormAssetRepository,
	service.NewAssetService,
	handler.NewAssetHandler,
)

// InitializeAssetHandler is the injector for AssetHandler and its dependencies.
func InitializeAssetHandler(db *gorm.DB, logger *zap.Logger, envRepo repository.EnvironmentRepository) (*handler.AssetHandler, error) {
	wire.Build(
		AssetSet,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for Service components
var ServiceSet = wire.NewSet(
	repository.NewGormServiceRepository,
	repository.NewGormServiceTypeRepository,
	service.NewServiceService,
	handler.NewServiceHandler,
)

// InitializeServiceHandler is the injector for ServiceHandler and its dependencies.
func InitializeServiceHandler(db *gorm.DB, logger *zap.Logger) (*handler.ServiceHandler, error) {
	wire.Build(
		ServiceSet,
	)
	return nil, nil // Wire will replace this
}

// InitializeServiceRepository is the injector for ServiceRepository.
func InitializeServiceRepository(db *gorm.DB) (repository.ServiceRepository, error) {
	wire.Build(
		repository.NewGormServiceRepository,
	)
	return nil, nil // Wire will replace this
}

// InitializeServiceTypeRepository is the injector for ServiceTypeRepository.
func InitializeServiceTypeRepository(db *gorm.DB) (repository.ServiceTypeRepository, error) {
	wire.Build(
		repository.NewGormServiceTypeRepository,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for service instance components
var ServiceInstanceSet = wire.NewSet(
	repository.NewServiceInstanceRepository,
	service.NewServiceInstanceService,
	handler.NewServiceInstanceHandler,
	// We need ServiceRepository and EnvironmentRepository for NewServiceInstanceService
	// Assuming they are provided directly to the injector or via other sets.
)

// InitializeServiceInstanceHandler is the injector for ServiceInstanceHandler.
func InitializeServiceInstanceHandler(
	db *gorm.DB,
	logger *zap.Logger,
	serviceRepo repository.ServiceRepository,
	envRepo repository.EnvironmentRepository,
) (*handler.ServiceInstanceHandler, error) {
	wire.Build(
		ServiceInstanceSet,
		// Provide serviceRepo and envRepo directly to this build context
		// wire.Value(serviceRepo), // This is incorrect for interfaces; they should be parameters or bound.
		// wire.Value(envRepo),
		// db and logger are already parameters to the injector func, so Wire can use them.
	)
	return nil, nil // Wire will replace this
}

// 环境组件的Provider Set和Initialize函数已在上方定义

// ProviderSet for business components
var BusinessSet = wire.NewSet(
	repository.NewBusinessRepository,
	service.NewBusinessService,
	handler.NewBusinessHandler,
)

// InitializeBusinessHandler is the injector for BusinessHandler and its dependencies.
func InitializeBusinessHandler(db *gorm.DB, logger *zap.Logger) (*handler.BusinessHandler, error) {
	wire.Build(
		BusinessSet,
	)
	return nil, nil // Wire will replace this
}

// ProviderSet for bug management components
var BugSet = wire.NewSet(
	repository.NewBugRepository,
	service.NewBugService,
	handler.NewBugHandler,
)

// InitializeBugHandler is the injector for BugHandler and its dependencies.
func InitializeBugHandler(db *gorm.DB, logger *zap.Logger) (*handler.BugHandler, error) {
	wire.Build(
		BugSet,
	)
	return nil, nil // Wire will replace this
}
