package router

import (
	"EffiPlat/backend/internal/handler"    // Corrected import path
	"EffiPlat/backend/internal/middleware" // Corrected import path

	// userHandler "EffiPlat/backend/internal/user" // Removed
	"net/http" // 引入 net/http 包

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置和返回 Gin 引擎
// 添加 jwtKey 参数
func SetupRouter(authHandler *handler.AuthHandler, userHdlr *handler.UserHandler, roleHdlr *handler.RoleHandler, permissionHdlr *handler.PermissionHandler, jwtKey []byte /*, etc. */) *gin.Engine {
	r := gin.Default()

	// 禁用自动重定向
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// 可以添加全局中间件，例如:
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())
	// r.Use(middleware.CORSMiddleware()) // 假设有 CORS 中间件

	// 健康检查 (可以放在根路径或 API 组外)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		authRoutes(v1, authHandler, jwtKey)              // 传递 jwtKey
		userRoutes(v1, userHdlr, jwtKey)                 // userHdlr is now *handler.UserHandler
		roleRoutes(v1, roleHdlr, permissionHdlr, jwtKey) // Add this line for role routes
		permissionRoutes(v1, permissionHdlr, jwtKey)     // Register permission routes
		// ... 其他路由组 ...
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

// authRoutes 注册认证相关的路由
// 添加 jwtKey 参数
func authRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler, jwtKey []byte) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		// 取消注释并添加 Logout 和 GetMe 路由
		auth.POST("/logout", middleware.JWTAuthMiddleware(jwtKey), authHandler.Logout) // 应用 JWT 中间件
		auth.GET("/me", middleware.JWTAuthMiddleware(jwtKey), authHandler.GetMe)       // 应用 JWT 中间件
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
func roleRoutes(rg *gin.RouterGroup, roleHdlr *handler.RoleHandler, permissionHdlr *handler.PermissionHandler, jwtKey []byte) {
	roles := rg.Group("/roles")
	roles.Use(middleware.JWTAuthMiddleware(jwtKey))
	{
		roles.GET("", roleHdlr.GetRoles)             // GET /api/v1/roles
		roles.POST("", roleHdlr.CreateRole)          // POST /api/v1/roles
		roles.GET(":roleId", roleHdlr.GetRoleByID)   // GET /api/v1/roles/{roleId}
		roles.PUT(":roleId", roleHdlr.UpdateRole)    // PUT /api/v1/roles/{roleId}
		roles.DELETE(":roleId", roleHdlr.DeleteRole) // DELETE /api/v1/roles/{roleId}

		// Route to get permissions for a specific role
		roles.GET(":roleId/permissions", permissionHdlr.GetPermissionsByRoleID)
	}
}

// permissionRoutes registers permission management related routes
func permissionRoutes(rg *gin.RouterGroup, permissionHdlr *handler.PermissionHandler, jwtKey []byte) {
	permissions := rg.Group("/permissions")
	permissions.Use(middleware.JWTAuthMiddleware(jwtKey)) // Apply authentication middleware
	{
		permissions.GET("", permissionHdlr.GetPermissions)                   // GET /api/v1/permissions
		permissions.POST("", permissionHdlr.CreatePermission)                // POST /api/v1/permissions
		permissions.GET(":permissionId", permissionHdlr.GetPermissionByID)   // GET /api/v1/permissions/{permissionId}
		permissions.PUT(":permissionId", permissionHdlr.UpdatePermission)    // PUT /api/v1/permissions/{permissionId}
		permissions.DELETE(":permissionId", permissionHdlr.DeletePermission) // DELETE /api/v1/permissions/{permissionId}

		// Add routes for managing role permissions - these handlers are in permissionHdlr
		permissions.POST("/roles/:roleId", permissionHdlr.AddPermissionsToRole)        // POST /api/v1/permissions/roles/{roleId}
		permissions.DELETE("/roles/:roleId", permissionHdlr.RemovePermissionsFromRole) // DELETE /api/v1/permissions/roles/{roleId}
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
