# 用户认证 API 设计方案

## 1. 设计目标

- 提供安全、标准化的用户认证接口（登录、获取当前用户、登出等）
- 支持 JWT 认证，便于前后端分离和后续扩展
- 便于 wire 依赖注入和分层架构实现

## 2. 路由与接口规范

### 2.1 登录

- **POST /auth/login**
- **请求体：**
  ```json
  {
    "email": "user@example.com",
    "password": "yourpassword"
  }
  ```
- **响应体 (成功):**
  ```json
  {
    "token": "jwt_token_string",
    "user": {
      "id": 1,
      "name": "张三",
      "email": "user@example.com"
      // 注意：实际响应中可能不包含所有用户字段，特别是 PasswordHash
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "error": "invalid credentials" // 或 "invalid request"
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com", 
    "password": "password"
  }'
  ```

### 2.2 获取当前用户

- **GET /auth/me**
- **请求头：**
  ```
  Authorization: Bearer <jwt_token>
  ```
- **响应体：**
  ```json
  {
    "id": 1,
    "name": "张三",
    "email": "user@example.com"
  }
  ```
- **实现状态**: 已完成。
- **实现说明**: 通过 JWT 中间件获取 Claims，直接返回其中的用户信息。
- **Curl 示例:**
  ```bash
  # 将 <your_jwt_token> 替换为登录后获取的实际 token
  curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <your_jwt_token>"
  ```

### 2.3 登出

- **POST /auth/logout**
- **请求头：**
  ```
  Authorization: Bearer <jwt_token>
  ```
- **响应体：**
  ```json
  {
    "message": "logout successful"
  }
  ```
- **实现状态**: 已完成 (基础实现)。
- **实现说明**: 当前实现仅返回成功消息，客户端负责丢弃 Token。未实现服务端 Token 黑名单机制。
- **Curl 示例:**
  ```bash
  # 将 <your_jwt_token> 替换为登录后获取的实际 token
  curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <your_jwt_token>"
  ```

## 3. 数据结构与安全方案

- 密码加密：bcrypt
- Token：JWT（github.com/golang-jwt/jwt/v5）
- JWT Claims 结构体定义
- 认证中间件：Gin 自定义 JWT 校验

## 4. 依赖注入与分层实现建议

- repository/service/handler/middleware 全部通过 wire 注入
- 目录结构建议：
  ```
  internal/
    handler/
      auth_handler.go
    service/
      auth_service.go
    repository/
      user_repository.go
    model/
      user.go
      auth.go
    middleware/
      jwt_middleware.go
  ```

## 5. 扩展与安全注意事项

- 后续可扩展 OAuth、第三方登录
- 支持 token 黑名单、刷新机制
- 防止暴力破解、加强日志审计 