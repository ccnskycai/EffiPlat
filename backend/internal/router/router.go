package router

import (
	"EffiPlat/backend/internal/handler"    // 引入 handler 包
	"EffiPlat/backend/internal/middleware" // 取消注释 middleware 包
	"net/http"                             // 引入 net/http 包

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置和返回 Gin 引擎
// 添加 jwtKey 参数
func SetupRouter(authHandler *handler.AuthHandler, jwtKey []byte /*, userHandler *handler.UserHandler, etc. */) *gin.Engine {
	r := gin.Default()

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
		authRoutes(v1, authHandler, jwtKey) // 传递 jwtKey
		// userRoutes(v1, userHandler) // 添加其他模块的路由
		// ... 其他路由组 ...
	}

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

// --- 可以为每个模块添加类似的路由注册函数 ---
/*
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
