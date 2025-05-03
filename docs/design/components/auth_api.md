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
- **响应体：**
  ```json
  {
    "token": "jwt_token_string",
    "user": {
      "id": 1,
      "name": "张三",
      "email": "user@example.com"
    }
  }
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

### 2.3 登出

- **POST /auth/logout**
- **请求头：**
  ```
  Authorization: Bearer <jwt_token>
  ```
- **响应体：**
  ```json
  {
    "message": "logout success"
  }
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