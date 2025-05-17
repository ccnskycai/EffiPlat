# 角色管理 API 设计方案

## 1. 设计目标

- 提供完整的角色管理接口（创建、查询、更新、删除等）
- 支持角色与权限的关联管理 (虽然代码中权限部分是 TODO，但在设计文档中应体现目标)
- 支持角色的状态管理 (如果适用，当前代码未体现)
- 遵循统一的 API 响应规范
- 便于 wire 依赖注入和分层架构实现

## 2. 路由与接口规范

所有接口均位于 `/api/v1/roles` 路径下，并使用统一的响应格式 `{ "code": 0, "message": "Success", "data": ... }` 或 `{ "code": <error_code>, "message": "<error_message>", "data": null }`。

### 2.1 获取角色列表

- **GET /roles**
- **请求参数：**
  ```
  page: 页码（默认1）
  pageSize: 每页数量（默认10）
  name: 按名称搜索 (模糊匹配)
  ```
- **响应体 (成功):**
  ```json
  {
    "code": 0,
    "message": "Success",
    "data": {
      "items": [
        {
          "id": 1,
          "name": "管理员",
          "description": "拥有所有权限",
          "createdAt": "2024-01-01T10:00:00Z",
          "updatedAt": "2024-01-01T10:00:00Z"
        },
        {
          "id": 2,
          "name": "普通用户",
          "description": "基本用户权限",
          "createdAt": "2024-01-01T10:05:00Z",
          "updatedAt": "2024-01-01T10:05:00Z"
        }
      ],
      "total": 2,
      "page": 1,
      "pageSize": 10
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "code": 50000,
    "message": "Failed to retrieve roles",
    "data": null
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json"
  ```

### 2.2 创建角色

- **POST /roles**
- **请求体：**
  ```json
  {
    "name": "新角色",
    "description": "这是新角色的描述",
    "permissionIds": [1, 3] // 可选：关联的权限ID列表
  }
  ```
- **响应体 (成功):**
  ```json
  {
    "code": 0,
    "message": "Success",
    "data": {
      "id": 3,
      "name": "新角色",
      "description": "这是新角色的描述",
      "createdAt": "2024-01-01T10:10:00Z",
      "updatedAt": "2024-01-01T10:10:00Z"
      // 注意：创建成功响应中可能不包含完整的权限列表，取决于后端实现
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "code": 40000,
    "message": "Invalid request payload: name is required",
    "data": null
  }
  // 或
  {
    "code": 40002,
    "message": "Validation error: Role name already exists",
    "data": null
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Role",
    "description": "Description of the new role",
    "permissionIds": [1, 2]
  }'
  ```

### 2.3 获取单个角色

- **GET /roles/{roleId}**
- **路径参数:** `roleId` (integer) - 角色ID
- **响应体 (成功):**
  ```json
  {
    "code": 0,
    "message": "Success",
    "data": {
      "id": 1,
      "name": "管理员",
      "description": "拥有所有权限",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z",
      "userCount": 5, // 关联的用户数量 (TODO 实现)
      "permissions": [ // 关联的权限列表 (TODO 实现)
        {
          "id": 101,
          "name": "user:read",
          "description": "查看用户信息"
        },
        {
          "id": 102,
          "name": "user:write",
          "description": "修改用户信息"
        }
      ]
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "code": 40000,
    "message": "Invalid role ID format",
    "data": null
  }
  // 或
  {
    "code": 40402,
    "message": "Role not found",
    "data": null
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X GET http://localhost:8080/api/v1/roles/1 \
  -H "Authorization: Bearer <your_jwt_token>"
  ```

### 2.4 更新角色

- **PUT /roles/{roleId}**
- **路径参数:** `roleId` (integer) - 角色ID
- **请求体：**
  ```json
  {
    "name": "更新后的角色名",
    "description": "更新后的描述",
    "permissionIds": [1, 4, 5] // 可选：新的关联权限ID列表
  }
  ```
- **响应体 (成功):**
  ```json
  {
    "code": 0,
    "message": "Success",
    "data": {
      "id": 1,
      "name": "更新后的角色名",
      "description": "更新后的描述",
      "createdAt": "2024-01-01T10:00:00Z", // 创建时间不变
      "updatedAt": "2024-01-01T10:15:00Z" // 更新为当前时间
      // 注意：更新成功响应中可能不包含完整的权限列表，取决于后端实现
    }
  }
  ```
- **响应体 (失败):**
  ```json
  {
    "code": 40000,
    "message": "Invalid request payload: name is required",
    "data": null
  }
  // 或
  {
    "code": 40000,
    "message": "Invalid role ID format",
    "data": null
  }
  // 或
  {
    "code": 40402,
    "message": "Role not found",
    "data": null
  }
   // 或
  {
    "code": 40002,
    "message": "Validation error: Role name already exists",
    "data": null
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X PUT http://localhost:8080/api/v1/roles/1 \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Role Name",
    "description": "This is the updated description",
    "permissionIds": [3]
  }'
  ```

### 2.5 删除角色

- **DELETE /roles/{roleId}**
- **路径参数:** `roleId` (integer) - 角色ID
- **响应体 (成功):**
  成功删除返回 HTTP 状态码 `204 No Content`，响应体为空。
  （与 User API 文档的成功删除返回 200 和 `{ status: "success", message: "...", data: null }` 略有不同，以实际代码实现为准，当前代码返回 204）
- **响应体 (失败):**
  ```json
  {
    "code": 40000,
    "message": "Invalid role ID format",
    "data": null
  }
  // 或
  {
    "code": 40402,
    "message": "Role not found",
    "data": null
  }
   // 或
  {
    "code": 40003,
    "message": "Role is assigned to users and cannot be deleted", // 角色被用户引用时
    "data": null
  }
  ```
- **Curl 示例:**
  ```bash
  curl -X DELETE http://localhost:8080/api/v1/roles/1 \
  -H "Authorization: Bearer <your_jwt_token>"
  ```

## 3. 数据结构与安全方案

- **Role Model:** 参照 `backend/internal/models/user.go` 中的 `Role` 结构体定义。
- **Permission Model:** 需要定义权限相关的 Model。
- **关联管理:** 使用 GORM 的 many2many 标签维护角色与用户、角色与权限的关系。
- **数据验证:** 使用 binding 标签进行请求体基础验证，Service 层进行业务逻辑验证（如名称唯一性）。
- **访问控制:** Role API 应该只允许具有特定权限的用户访问（例如，只有管理员可以进行增删改查）。通过 JWT 和 RBAC 中间件实现。

## 4. 依赖注入与分层实现

- repository/service/handler 全部通过 wire 注入。
- 目录结构：
  ```
  internal/
    handler/
      role_handler.go
    service/
      role_service.go
    repository/
      role_repository.go
    model/
      role_models.go // 用于请求/响应结构
      user.go // 包含 Role Model 定义
      permission_models.go // 权限相关 Model (TODO)
  ```

## 5. 扩展与安全注意事项

- **权限管理:** 需要完善权限相关的 Model、Repository、Service 和 Handler。
- **用户关联:** 在删除角色时，需要处理与该角色关联的用户。目前的 handler 依赖 service 层实现此逻辑（例如，如果角色仍被用户引用则禁止删除）。
- **操作日志:** 记录敏感的角色管理操作。
- **接口限流:** 防止恶意请求。
- **字段过滤:** 避免在 GET 请求中暴露不必要的敏感信息。 