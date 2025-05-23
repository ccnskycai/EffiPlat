# 用户管理 API 设计方案

## 1. 设计目标

- 提供完整的用户管理接口（创建、查询、更新、删除等）
- 支持基于角色的访问控制（RBAC）
- 支持用户状态管理和部门管理
- 便于 wire 依赖注入和分层架构实现

## 2. 路由与接口规范

### 2.1 获取用户列表

- **GET /users**
- **请求参数：**
  ```
  page: 页码（默认1）
  pageSize: 每页数量（默认10）
  name: 按名称搜索
  email: 按邮箱搜索
  status: 按状态筛选
  sortBy: 排序字段
  order: 排序方向（asc/desc）
  ```
- **响应体 (成功):**
  ```json
  {
    "status": "success",
    "message": "Users retrieved successfully",
    "data": {
      "items": [
        {
          "id": 1,
          "name": "张三",
          "email": "zhangsan@example.com",
          "department": "研发部",
          "status": "active",
          "roles": [
            {
              "id": 1,
              "name": "admin"
            }
          ],
          "createdAt": "2024-03-14T10:00:00Z",
          "updatedAt": "2024-03-14T10:00:00Z"
        }
      ],
      "total": 100,
      "page": 1,
      "pageSize": 10
    }
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json"
  ```

### 2.2 创建用户

- **POST /users**
- **请求体：**
  ```json
  {
    "name": "张三",
    "email": "zhangsan@example.com",
    "password": "securepassword123",
    "department": "研发部",
    "roles": [1, 2] // 角色ID列表
  }
  ```
- **响应体 (成功):**
  ```json
  {
    "status": "success",
    "message": "User created successfully",
    "data": {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "department": "研发部",
      "status": "active",
      "roles": [
        {
          "id": 1,
          "name": "admin"
        },
        {
          "id": 2,
          "name": "developer"
        }
      ],
      "createdAt": "2024-03-14T10:00:00Z",
      "updatedAt": "2024-03-14T10:00:00Z"
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "status": "error",
    "message": "Failed to create user",
    "errors": "Email already exists"
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "password": "securepassword123",
    "department": "研发部",
    "roles": [1, 2]
  }'
  ```

### 2.3 获取单个用户

- **GET /users/{userId}**
- **响应体 (成功):**
  ```json
  {
    "status": "success",
    "message": "User retrieved successfully",
    "data": {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "department": "研发部",
      "status": "active",
      "roles": [
        {
          "id": 1,
          "name": "admin"
        }
      ],
      "createdAt": "2024-03-14T10:00:00Z",
      "updatedAt": "2024-03-14T10:00:00Z"
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "status": "error",
    "message": "User not found",
    "errors": "User with ID 1 not found"
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json"
  ```

### 2.4 更新用户

- **PUT /users/{userId}**
- **请求体：**
  ```json
  {
    "name": "张三（更新）",
    "department": "产品部",
    "status": "inactive",
    "roles": [2, 3]
  }
  ```
- **响应体 (成功):**
  ```json
  {
    "status": "success",
    "message": "User updated successfully",
    "data": {
      "id": 1,
      "name": "张三（更新）",
      "email": "zhangsan@example.com",
      "department": "产品部",
      "status": "inactive",
      "roles": [
        {
          "id": 2,
          "name": "developer"
        },
        {
          "id": 3,
          "name": "tester"
        }
      ],
      "createdAt": "2024-03-14T10:00:00Z",
      "updatedAt": "2024-03-14T10:30:00Z"
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "status": "error",
    "message": "Failed to update user",
    "errors": "User not found or update failed"
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三（更新）",
    "department": "产品部",
    "status": "inactive",
    "roles": [2, 3]
  }'
  ```

### 2.5 删除用户

- **DELETE /users/{userId}**
- **响应体 (成功):**
  ```json
  {
    "status": "success",
    "message": "User deleted successfully",
    "data": null
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "status": "error",
    "message": "Failed to delete user",
    "errors": "User not found or delete failed"
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X DELETE http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json"
  ```

### 2.6 为用户分配角色

- **POST /users/{userId}/roles**
- **路径参数:** `userId` (integer) - 用户ID
- **请求体：**
  ```json
  {
    "role_ids": [1, 2] // 要分配给用户的角色ID列表
  }
  ```
- **响应体 (成功 - HTTP 200):**
  ```json
  {
    "code": 0,
    "message": "Roles assigned successfully",
    "data": null
  }
  ```
- **响应体 (失败):**
  - HTTP 400 (Bad Request): 如果用户ID格式无效、角色ID无效或请求体格式错误。
    ```json
    {
      "code": 40000, // 示例错误码
      "message": "Invalid role ID: 999",
      "data": null
    }
    ```
  - HTTP 404 (Not Found): 如果用户不存在。
    ```json
    {
      "code": 40400, // 示例错误码
      "message": "User not found",
      "data": null
    }
    ```
- **Curl 示例:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/users/1/roles \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [1, 2]
  }'
  ```

### 2.7 从用户移除角色

- **DELETE /users/{userId}/roles**
- **路径参数:** `userId` (integer) - 用户ID
- **请求体：**
  ```json
  {
    "role_ids": [1] // 要从用户移除的角色ID列表
  }
  ```
- **响应体 (成功 - HTTP 200):**
  ```json
  {
    "code": 0,
    "message": "Roles removed successfully",
    "data": null
  }
  ```
- **响应体 (失败):**
  - HTTP 400 (Bad Request): 如果用户ID格式无效、角色ID无效或请求体格式错误。
  - HTTP 404 (Not Found): 如果用户不存在。
- **Curl 示例:**
  ```bash
  curl -X DELETE http://localhost:8080/api/v1/users/1/roles \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [1]
  }'
  ```

## 3. 数据结构与安全方案

- 密码处理：使用 bcrypt 进行密码哈希
- 角色管理：基于 RBAC 模型
- 状态管理：支持 active/inactive/pending 等状态
- 数据验证：使用 validator 进行请求体验证

## 4. 依赖注入与分层实现

- repository/service/handler 全部通过 wire 注入
- 目录结构：
  ```
  internal/
    handler/
      user_handler.go
    service/
      user_service.go
    repository/
      user_repository.go
    model/
      user.go
  ```

## 5. 扩展与安全注意事项

- 支持用户状态变更审计
- 支持角色变更审计
- 防止越权访问（用户只能修改自己的信息，管理员可以修改所有用户）
- 密码策略（长度、复杂度等）
- 邮箱验证机制
- 用户操作日志记录
