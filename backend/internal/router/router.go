package router

import (
	"EffiPlat/backend/internal/handler" // Unified import path for all handlers
	"EffiPlat/backend/internal/middleware"           // Corrected import path

	// userHandler "EffiPlat/backend/internal/user" // Removed
	"net/http" // 引入 net/http 包
	"regexp"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var alphanumDashRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")

// validateAlphanumDash implements validator.Func for 'alphanumdash' tag
func validateAlphanumDash(fl validator.FieldLevel) bool {
	return alphanumDashRegex.MatchString(fl.Field().String())
}

// SetupRouter 配置和返回 Gin 引擎
// 添加 jwtKey 参数
func SetupRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	responsibilityHandler *handler.ResponsibilityHandler,
	responsibilityGroupHandler *handler.ResponsibilityGroupHandler,
	environmentHandler *handler.EnvironmentHandler,
	assetHandler *handler.AssetHandler,
	serviceHandler *handler.ServiceHandler,
	serviceInstanceHandler *handler.ServiceInstanceHandler, // Added ServiceInstanceHandler
	businessHandler *handler.BusinessHandler, // Added BusinessHandler
	bugHandler *handler.BugHandler, // Added BugHandler
	jwtKey []byte,
) *gin.Engine {
	r := gin.Default()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("alphanumdash", validateAlphanumDash)
	}

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Logger middleware (Gin's default logger is quite good)
	// For custom logging, you can add r.Use(middleware.LoggingMiddleware(logger)) here if you have one

	// 禁用自动重定向
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// 健康检查 (可以放在根路径或 API 组外)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Public routes (e.g., login, register if it were public)
	apiV1Public := r.Group("/api/v1")
	{
		// Only login should be public within auth group
		publicAuth := apiV1Public.Group("/auth")
		{
			publicAuth.POST("/login", authHandler.Login)
		}
		// If user registration was public, it would be here
		// e.g., apiV1Public.POST("/register", userHandler.RegisterUser) // Example, if RegisterUser exists and is public
	}

	// Authenticated routes
	apiV1Authenticated := r.Group("/api/v1")
	apiV1Authenticated.Use(middleware.JWTAuthMiddleware(jwtKey))
	{
		// Authenticated Auth routes (me, logout)
		authAuth := apiV1Authenticated.Group("/auth")
		{
			authAuth.GET("/me", authHandler.GetMe)
			authAuth.POST("/logout", authHandler.Logout)
		}

		// User routes (already includes CRUD for users)
		userRoutes := apiV1Authenticated.Group("/users")
		{
			userRoutes.GET("", userHandler.GetUsers)
			userRoutes.POST("", userHandler.CreateUser) // CreateUser might be admin-only or public depending on policy
			userRoutes.GET("/:userId", userHandler.GetUserByID)
			userRoutes.PUT("/:userId", userHandler.UpdateUser)
			userRoutes.DELETE("/:userId", userHandler.DeleteUser)

			// Routes for assigning/removing roles to/from a user
			userRoutes.POST("/:userId/roles", userHandler.AssignRolesToUser)
			userRoutes.DELETE("/:userId/roles", userHandler.RemoveRolesFromUser)
		}

		// Role routes
		roleRoutes(apiV1Authenticated.Group("/roles"), roleHandler, permissionHandler) // permissionHandler was added for /roles/{roleId}/permissions

		// Permission routes (already includes CRUD for permissions and associating them with roles)
		permissionRoutes(apiV1Authenticated.Group("/permissions"), permissionHandler)

		// Responsibility routes
		responsibilityRoutes(apiV1Authenticated.Group("/responsibilities"), responsibilityHandler)

		// Responsibility Group routes
		responsibilityGroupRoutes(apiV1Authenticated.Group("/responsibility-groups"), responsibilityGroupHandler)

		// Environment routes
		environmentRoutes(apiV1Authenticated.Group("/environments"), environmentHandler)

		// Asset routes
		assetRoutes(apiV1Authenticated.Group("/assets"), assetHandler)

		// ServiceType and Service routes
		serviceTypeRoutes(apiV1Authenticated.Group("/service-types"), serviceHandler)
		serviceRoutes(apiV1Authenticated.Group("/services"), serviceHandler)

		// Service Instance routes
		serviceInstanceGroup := apiV1Authenticated.Group("/service-instances")
		{
			serviceInstanceGroup.POST("", serviceInstanceHandler.CreateServiceInstance)
			serviceInstanceGroup.GET("", serviceInstanceHandler.ListServiceInstances)
			serviceInstanceGroup.GET("/:instanceId", serviceInstanceHandler.GetServiceInstance)
			serviceInstanceGroup.PUT("/:instanceId", serviceInstanceHandler.UpdateServiceInstance)
			serviceInstanceGroup.DELETE("/:instanceId", serviceInstanceHandler.DeleteServiceInstance)
		}

		// Business routes
		businessRoutes(apiV1Authenticated.Group("/businesses"), businessHandler)

		// Bug routes
		bugRoutes(apiV1Authenticated.Group("/bugs"), bugHandler)
	}

	// 处理404路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"path":  c.Request.URL.Path,
		})
	})

	// 可选：为根路径添加一个简单响应
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "EffiPlat Backend is running!")
	})

	return r
}

// authRoutes 注册认证相关的路由 - NOW ONLY FOR LOGIN
// Remove jwtKey parameter as it's no longer needed for a public login route
func authRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler) {
	// This function is now effectively reduced or could be inlined above
	// For clarity, let's assume it's still called for login only.
	auth := rg.Group("/auth") // This was previously apiV1Public, so path is /api/v1/auth
	{
		auth.POST("/login", authHandler.Login)
		// GET /me and POST /logout are now under apiV1Authenticated
	}
}

// userRoutes 注册用户管理相关的路由
func userRoutes(rg *gin.RouterGroup, userHdlr *handler.UserHandler, jwtKey []byte) {
	users := rg.Group("/users")
	users.Use(middleware.JWTAuthMiddleware(jwtKey))
	{
		users.GET("", userHdlr.GetUsers)             // GET /api/v1/users
		users.POST("", userHdlr.CreateUser)          // POST /api/v1/users
		users.GET(":userId", userHdlr.GetUserByID)   // GET /api/v1/users/{userId}
		users.PUT(":userId", userHdlr.UpdateUser)    // PUT /api/v1/users/{userId}
		users.DELETE(":userId", userHdlr.DeleteUser) // DELETE /api/v1/users/{userId}
	}
}

// roleRoutes 注册角色管理相关的路由
func roleRoutes(rg *gin.RouterGroup, roleHdlr *handler.RoleHandler, permissionHdlr *handler.PermissionHandler) {
	{
		rg.GET("", roleHdlr.GetRoles)             // GET /api/v1/roles
		rg.POST("", roleHdlr.CreateRole)          // POST /api/v1/roles
		rg.GET(":roleId", roleHdlr.GetRoleByID)   // GET /api/v1/roles/{roleId}
		rg.PUT(":roleId", roleHdlr.UpdateRole)    // PUT /api/v1/roles/{roleId}
		rg.DELETE(":roleId", roleHdlr.DeleteRole) // DELETE /api/v1/roles/{roleId}

		// Route to get permissions for a specific role
		rg.GET(":roleId/permissions", permissionHdlr.GetPermissionsByRoleID)
	}
}

// permissionRoutes registers permission management related routes
func permissionRoutes(rg *gin.RouterGroup, permissionHdlr *handler.PermissionHandler) {
	{
		rg.GET("", permissionHdlr.GetPermissions)                   // GET /api/v1/permissions
		rg.POST("", permissionHdlr.CreatePermission)                // POST /api/v1/permissions
		rg.GET(":permissionId", permissionHdlr.GetPermissionByID)   // GET /api/v1/permissions/{permissionId}
		rg.PUT(":permissionId", permissionHdlr.UpdatePermission)    // PUT /api/v1/permissions/{permissionId}
		rg.DELETE(":permissionId", permissionHdlr.DeletePermission) // DELETE /api/v1/permissions/{permissionId}

		// Add routes for managing role permissions - these handlers are in permissionHdlr
		rg.POST("/roles/:roleId", permissionHdlr.AddPermissionsToRole)        // POST /api/v1/permissions/roles/{roleId}
		rg.DELETE("/roles/:roleId", permissionHdlr.RemovePermissionsFromRole) // DELETE /api/v1/permissions/roles/{roleId}
	}
}

// responsibilityRoutes 注册职责管理相关的路由
func responsibilityRoutes(rg *gin.RouterGroup, hdlr *handler.ResponsibilityHandler) {
	{
		rg.POST("", hdlr.CreateResponsibility)                     // POST /api/v1/responsibilities
		rg.GET("", hdlr.GetResponsibilities)                       // GET /api/v1/responsibilities
		rg.GET("/:responsibilityId", hdlr.GetResponsibilityByID)   // GET /api/v1/responsibilities/{responsibilityId}
		rg.PUT("/:responsibilityId", hdlr.UpdateResponsibility)    // PUT /api/v1/responsibilities/{responsibilityId}
		rg.DELETE("/:responsibilityId", hdlr.DeleteResponsibility) // DELETE /api/v1/responsibilities/{responsibilityId}
	}
}

// responsibilityGroupRoutes 注册职责组管理相关的路由
func responsibilityGroupRoutes(rg *gin.RouterGroup, hdlr *handler.ResponsibilityGroupHandler) {
	{
		rg.POST("", hdlr.CreateResponsibilityGroup)            // POST /api/v1/responsibility-groups
		rg.GET("", hdlr.GetResponsibilityGroups)               // GET /api/v1/responsibility-groups
		rg.GET("/:groupId", hdlr.GetResponsibilityGroupByID)   // GET /api/v1/responsibility-groups/{groupId}
		rg.PUT("/:groupId", hdlr.UpdateResponsibilityGroup)    // PUT /api/v1/responsibility-groups/{groupId}
		rg.DELETE("/:groupId", hdlr.DeleteResponsibilityGroup) // DELETE /api/v1/responsibility-groups/{groupId}

		// Routes for managing responsibilities within a group
		rg.POST("/:groupId/responsibilities/:responsibilityId", hdlr.AddResponsibilityToGroup)        // POST /api/v1/responsibility-groups/{groupId}/responsibilities/{responsibilityId}
		rg.DELETE("/:groupId/responsibilities/:responsibilityId", hdlr.RemoveResponsibilityFromGroup) // DELETE /api/v1/responsibility-groups/{groupId}/responsibilities/{responsibilityId}
	}
}

// environmentRoutes 注册环境管理相关的路由
func environmentRoutes(rg *gin.RouterGroup, hdlr *handler.EnvironmentHandler) {
	{
		rg.POST("", hdlr.CreateEnvironment)              // POST /api/v1/environments
		rg.GET("", hdlr.GetEnvironments)                 // GET /api/v1/environments
		rg.GET("/:id", hdlr.GetEnvironmentByID)          // GET /api/v1/environments/{id}
		rg.GET("/slug/:slug", hdlr.GetEnvironmentBySlug) // GET /api/v1/environments/slug/{slug}
		rg.PUT("/:id", hdlr.UpdateEnvironment)           // PUT /api/v1/environments/{id}
		rg.DELETE("/:id", hdlr.DeleteEnvironment)        // DELETE /api/v1/environments/{id}
	}
}

// assetRoutes 注册资产管理相关的路由
func assetRoutes(rg *gin.RouterGroup, hdlr *handler.AssetHandler) {
	{
		rg.POST("", hdlr.CreateAsset)       // POST /api/v1/assets
		rg.GET("", hdlr.ListAssets)         // GET /api/v1/assets
		rg.GET("/:id", hdlr.GetAssetByID)   // GET /api/v1/assets/{id}
		rg.PUT("/:id", hdlr.UpdateAsset)    // PUT /api/v1/assets/{id}
		rg.DELETE("/:id", hdlr.DeleteAsset) // DELETE /api/v1/assets/{id}
	}
}

// serviceTypeRoutes 注册服务类型管理相关的路由
func serviceTypeRoutes(rg *gin.RouterGroup, hdlr *handler.ServiceHandler) {
	{
		rg.POST("", hdlr.CreateServiceType)       // POST /api/v1/service-types
		rg.GET("", hdlr.ListServiceTypes)         // GET /api/v1/service-types
		rg.GET("/:id", hdlr.GetServiceTypeByID)   // GET /api/v1/service-types/{id}
		rg.PUT("/:id", hdlr.UpdateServiceType)    // PUT /api/v1/service-types/{id}
		rg.DELETE("/:id", hdlr.DeleteServiceType) // DELETE /api/v1/service-types/{id}
	}
}

// serviceRoutes 注册服务管理相关的路由
func serviceRoutes(rg *gin.RouterGroup, hdlr *handler.ServiceHandler) {
	{
		rg.POST("", hdlr.CreateService)       // POST /api/v1/services
		rg.GET("", hdlr.ListServices)         // GET /api/v1/services
		rg.GET("/:id", hdlr.GetServiceByID)   // GET /api/v1/services/{id}
		rg.PUT("/:id", hdlr.UpdateService)    // PUT /api/v1/services/{id}
		rg.DELETE("/:id", hdlr.DeleteService) // DELETE /api/v1/services/{id}
	}
}

// businessRoutes 注册业务管理相关的路由
func businessRoutes(rg *gin.RouterGroup, businessHdlr *handler.BusinessHandler) {
	{
		rg.POST("", businessHdlr.CreateBusiness)               // POST /api/v1/businesses
		rg.GET("", businessHdlr.ListBusinesses)                // GET /api/v1/businesses
		rg.GET("/:businessId", businessHdlr.GetBusinessByID)   // GET /api/v1/businesses/{businessId}
		rg.PUT("/:businessId", businessHdlr.UpdateBusiness)    // PUT /api/v1/businesses/{businessId}
		rg.DELETE("/:businessId", businessHdlr.DeleteBusiness) // DELETE /api/v1/businesses/{businessId}
	}
}

// bugRoutes 注册bug管理相关的路由
func bugRoutes(rg *gin.RouterGroup, bugHdlr *handler.BugHandler) {
	{
		rg.POST("", bugHdlr.CreateBug)       // POST /api/v1/bugs
		rg.GET("", bugHdlr.ListBugs)         // GET /api/v1/bugs
		rg.GET("/:id", bugHdlr.GetBugByID)   // GET /api/v1/bugs/{id}
		rg.PUT("/:id", bugHdlr.UpdateBug)    // PUT /api/v1/bugs/{id}
		rg.DELETE("/:id", bugHdlr.DeleteBug) // DELETE /api/v1/bugs/{id}
	}
}

/*
// 原有的 userRoutes 示例可以删除或保留作为参考
func userRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler) {
    users := rg.Group("/users")
    // users.Use(middleware.JWTAuthMiddleware(jwtKey)) // 应用认证中间件
    {
        users.GET("/", userHandler.ListUsers)
        users.POST("/", userHandler.CreateUser)
        users.GET("/:userId", userHandler.GetUser)
        // ... 其他用户相关路由
    }
}
*/
