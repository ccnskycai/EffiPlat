# API 设计规范

## 1. 设计原则

*   **风格**: 遵循 RESTful 设计风格。
*   **传输安全**: 所有 API 请求必须通过 HTTPS 进行以确保传输安全。
*   **数据格式**: 请求体和响应体统一使用 JSON 格式。
*   **命名**: URL 路径使用小写字母和连字符 (`-`)，资源名使用复数形式 (如 `/users`, `/environments`)。
*   **HTTP 方法**: 合理使用 HTTP 方法表达操作意图：
    *   `GET`: 获取资源。
    *   `POST`: 创建资源。
    *   `PUT`: 完整更新资源。
    *   `PATCH`: 部分更新资源。
    *   `DELETE`: 删除资源。
*   **状态码**: 使用标准的 HTTP 状态码表示请求结果：
    *   `200 OK`: 请求成功 (GET, PUT, PATCH)。
    *   `201 Created`: 资源创建成功 (POST)。
    *   `204 No Content`: 请求成功，但无返回内容 (DELETE)。
    *   `400 Bad Request`: 客户端请求错误（如参数校验失败）。
    *   `401 Unauthorized`: 未认证。
    *   `403 Forbidden`: 已认证但无权限访问。
    *   `404 Not Found`: 请求的资源不存在。
    *   `500 Internal Server Error`: 服务端内部错误。
*   **统一响应格式**: 设计统一的 JSON 响应结构，包含状态码/业务码、消息和数据。
    ```json
    // 成功响应
    {
      "code": 0, // 0 表示成功
      "message": "Success",
      "data": { ... } // 或 [...] 或 null
    }
    // 错误响应
    {
      "code": 4001, // 非 0 自定义业务错误码 或 标准 HTTP 状态码的扩展
      "message": "Invalid input parameter: email format error",
      "data": null
    }
    // 分页响应示例 (包装在 data 中)
    {
      "code": 0,
      "message": "Success",
      "data": {
        "items": [ ... ], // 当前页数据列表
        "total": 150,     // 总记录数
        "page": 1,        // 当前页码
        "pageSize": 10    // 每页数量
      }
    }
    ```

## 2. 认证与授权

*   **认证**: 初步考虑使用基于 Token 的认证方式 (如 JWT)。用户登录成功后，服务端颁发 Token，客户端后续请求在 Header 中携带 Token (`Authorization: Bearer <token>`)。
*   **授权**: 基于 RBAC 模型。服务端在处理请求时，根据用户角色和预定义的权限规则，判断用户是否有权执行该操作。

## 3. 版本管理

*   API 应考虑版本管理，以支持未来的非兼容性变更。初步考虑在 URL 中加入版本号 (如 `/api/v1/...`)。

## 4. 核心 API 端点 (V1.0 关注点)

根据 V1.0 计划 (`docs/requirements/execution_plan.md`)，定义核心资源的操作接口

*   **用户认证** (`/api/v1/auth`):
    *   `POST /login`
        *   **Request Body**: `{ "email": "user@example.com", "password": "your_password" }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Login successful", "data": { "token": "jwt_token_string", "userId": 1, "name": "John Doe", "roles": ["admin", "developer"] } }`
        *   **Response (Error - 401 Unauthorized)**: `{ "code": 40101, "message": "Invalid credentials", "data": null }`
    *   `POST /logout` (需要认证)
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Logout successful", "data": null }`

*   **用户管理** (`/api/v1/users`):
    *   `GET /` (需要认证, 可能需要特定权限)
        *   **Query Params**: `page=1`, `pageSize=10`, `sortBy=createdAt`, `order=desc`, `status=active`, `email=...`, `name=...`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, "message": "Success", "data": { "items": [ { "id": 1, "name": "John Doe", "email": "john@example.com", "department": "R&D", "status": "active", "createdAt": "...", "updatedAt": "..." } ], "total": 1, "page": 1, "pageSize": 10 } }`
    *   `POST /` (需要认证, 可能需要特定权限)
        *   **Request Body**: `{ "name": "Jane Doe", "email": "jane@example.com", "password": "new_password", "department": "QA", "roles": [2, 3] }` (roles 是 role ID 列表)
        *   **Response (Success - 201 Created)**: `{ "code": 0, "message": "User created successfully", "data": { "id": 2, "name": "Jane Doe", ... } }`
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40001, "message": "Validation error: Email already exists", "data": null }`
    *   `GET /{userId}` (需要认证, 用户可查自己信息，管理员可查他人)
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "name": "John Doe",
                "email": "john@example.com",
                "department": "R&D",
                "status": "active",
                "createdAt": "...",
                "updatedAt": "...",
                "roles": [
                  { "id": 1, "name": "admin" }
                ],
                "assignedResponsibilities": [
                  { "responsibilityId": 1, "responsibilityName": "Database Administrator", "isPrimary": true },
                  { "responsibilityId": 5, "responsibilityName": "Service Owner - Auth", "isPrimary": false }
                ]
              }
            }
            ```
        *   **Response (Error - 404 Not Found)**: `{ "code": 40401, "message": "User not found", "data": null }`
    *   `PUT /{userId}` (需要认证, 用户可改自己信息，管理员可改他人)
        *   **Request Body**: `{ "name": "Jane Smith", "department": "QA-Lead", "status": "active" }` (不允许修改密码和邮箱，除非有特定接口)
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "User updated successfully", "data": { "id": 2, "name": "Jane Smith", ... } }`
    *   `DELETE /{userId}` (需要认证, 通常仅管理员)
        *   **Response (Success - 204 No Content)**: (无响应体)
        *   **Response (Error - 403 Forbidden)**: `{ "code": 40301, "message": "Permission denied", "data": null }`

*   **角色与权限管理** (`/api/v1/roles`, `/api/v1/permissions`) (管理员接口):
    *   `GET /roles`
        *   **Query Params**: `page=1`, `pageSize=20`, `name=...` (按名称模糊搜索)
        *   **Response (Success - 200 OK)**: (分页结构)
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "items": [
                  {
                    "id": 1,
                    "name": "Administrator",
                    "description": "Manages all system settings",
                    "createdAt": "...",
                    "updatedAt": "..."
                  },
                  {
                    "id": 2,
                    "name": "Developer",
                    "description": "Manages services and deployments",
                    "createdAt": "...",
                    "updatedAt": "..."
                  }
                ],
                "total": 5,
                "page": 1,
                "pageSize": 20
              }
            }
            ```
    *   `POST /roles`
        *   **Request Body**:
            ```json
            {
              "name": "QA Tester",
              "description": "Responsible for testing and bug reporting",
              "permissionIds": [5, 8, 12] // 关联的权限 ID 列表
            }
            ```
        *   **Response (Success - 201 Created)**:
            ```json
            {
              "code": 0,
              "message": "Role created successfully",
              "data": {
                "id": 3,
                "name": "QA Tester",
                "description": "Responsible for testing and bug reporting",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40002, "message": "Validation error: Role name already exists", "data": null }`
    *   `GET /roles/{roleId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 2,
                "name": "Developer",
                "description": "Manages services and deployments",
                "createdAt": "...",
                "updatedAt": "...",
                "userCount": 25,
                "permissions": [
                  { "id": 5, "name": "view_service", "description": "..." },
                  { "id": 8, "name": "manage_bug", "description": "..." }
                ]
              }
            }
            ```
        *   **Response (Error - 404 Not Found)**: `{ "code": 40402, "message": "Role not found", "data": null }`
    *   `PUT /roles/{roleId}`
        *   **Request Body**:
            ```json
            {
              "name": "Senior Developer",
              "description": "Manages services, deployments, and environments",
              "permissionIds": [5, 8, 12, 15] // 更新后的权限 ID 列表 (全量更新)
            }
            ```
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Role updated successfully",
              "data": {
                "id": 2,
                "name": "Senior Developer",
                "description": "Manages services, deployments, and environments",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
    *   `DELETE /roles/{roleId}`
        *   **Response (Success - 204 No Content)**: (无响应体)
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40003, "message": "Cannot delete role: Role is assigned to users", "data": null }` (如果有关联用户)
    *   `GET /roles/{roleId}/permissions`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": [
                 { "id": 5, "name": "view_service", "description": "..." },
                 { "id": 8, "name": "manage_bug", "description": "..." }
              ]
            }
            ```
    *   `POST /roles/{roleId}/permissions` (Assign permissions)
        *   **Request Body**:
            ```json
            {
              "permissionIds": [12, 15] // 要新增分配给该角色的权限 ID 列表
            }
            ```
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Permissions assigned successfully", "data": null }`
    *   `DELETE /roles/{roleId}/permissions/{permissionId}` (Revoke permission)
        *   **Response (Success - 204 No Content)**: (无响应体)
    *   `GET /permissions` (List all available permissions, 可能需要分页)
        *   **Query Params**: `page=1`, `pageSize=50`, `name=...`
        *   **Response (Success - 200 OK)**: (分页结构)
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "items": [
                  { "id": 1, "name": "manage_user", "description": "Allows creating, updating, deleting users" },
                  { "id": 2, "name": "view_user", "description": "Allows viewing user information" },
                  { "id": 5, "name": "view_service", "description": "Allows viewing service details" },
                  // ... 其他权限
                ],
                "total": 30,
                "page": 1,
                "pageSize": 50
              }
            }
            ```

*   **职责与职责组管理** (`/api/v1/responsibilities`, `/api/v1/responsibility-groups`):
    *   `GET /responsibilities`
        *   **Query Params**: `page=1`, `pageSize=20`, `name=...` (按名称模糊搜索)
        *   **Response (Success - 200 OK)**: (分页结构)
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "items": [
                  {
                    "id": 1,
                    "name": "Database Administrator",
                    "description": "Manages database schemas and performance",
                    "createdAt": "...",
                    "updatedAt": "..."
                  },
                  {
                    "id": 2,
                    "name": "Frontend Lead",
                    "description": "Leads the frontend development team",
                    "createdAt": "...",
                    "updatedAt": "..."
                  }
                ],
                "total": 15,
                "page": 1,
                "pageSize": 20
              }
            }
            ```
    *   `POST /responsibilities`
        *   **Request Body**:
            ```json
            {
              "name": "Service Owner - User Service",
              "description": "Responsible for the User Service lifecycle"
            }
            ```
        *   **Response (Success - 201 Created)**:
            ```json
            {
              "code": 0,
              "message": "Responsibility created successfully",
              "data": {
                "id": 3,
                "name": "Service Owner - User Service",
                "description": "Responsible for the User Service lifecycle",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40004, "message": "Validation error: Responsibility name already exists", "data": null }`
    *   `GET /responsibilities/{respId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "name": "Database Administrator",
                "description": "Manages database schemas and performance",
                "createdAt": "...",
                "updatedAt": "...",
                "assignedUsers": [
                    { "userId": 5, "name": "Alice", "isPrimary": true },
                    { "userId": 8, "name": "Bob", "isPrimary": false }
                ]
              }
            }
            ```
        *   **Response (Error - 404 Not Found)**: `{ "code": 40403, "message": "Responsibility not found", "data": null }`
    *   `PUT /responsibilities/{respId}`
        *   **Request Body**:
            ```json
            {
              "name": "DBA Lead",
              "description": "Leads the Database Administration team"
            }
            ```
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Responsibility updated successfully",
              "data": {
                "id": 1,
                "name": "DBA Lead",
                "description": "Leads the Database Administration team",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
    *   `DELETE /responsibilities/{respId}`
        *   **Response (Success - 204 No Content)**: (无响应体)
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40005, "message": "Cannot delete responsibility: It is assigned to users", "data": null }`
    *   `GET /responsibility-groups` (获取职责分配列表)
        *   **Query Params**: `page=1`, `pageSize=20`, `responsibilityId={respId}`, `userId={userId}`
        *   **Response (Success - 200 OK)**: (分页结构)
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "items": [
                  {
                    "id": 101,
                    "responsibility": {
                       "id": 1,
                       "name": "Database Administrator"
                    },
                    "user": {
                       "id": 5,
                       "name": "Alice"
                    },
                    "isPrimary": true,
                    "assignedAt": "..."
                  },
                  {
                    "id": 102,
                    "responsibility": {
                       "id": 1,
                       "name": "Database Administrator"
                    },
                    "user": {
                       "id": 8,
                       "name": "Bob"
                    },
                    "isPrimary": false,
                    "assignedAt": "..."
                  }
                ],
                "total": 2,
                "page": 1,
                "pageSize": 20
              }
            }
            ```
    *   `POST /responsibility-groups` (分配用户到职责)
        *   **Request Body**:
            ```json
            {
              "responsibilityId": 1,
              "userId": 5,
              "isPrimary": true
            }
            ```
        *   **Response (Success - 201 Created)**:
            ```json
            {
              "code": 0,
              "message": "User assigned to responsibility successfully",
              "data": {
                 "id": 103,
                 "responsibilityId": 1,
                 "userId": 5,
                 "isPrimary": true,
                 "assignedAt": "..."
              }
            }
            ```
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40006, "message": "Assignment error: User already assigned or invalid ID", "data": null }`
    *   `PUT /responsibility-groups/{groupId}` (修改分配信息，例如更换主要负责人)
        *   **Request Body**: `{ "isPrimary": false }`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Responsibility assignment updated",
              "data": {
                 "id": 101,
                 "responsibilityId": 1,
                 "userId": 5,
                 "isPrimary": false,
                 "assignedAt": "..."
              }
            }
            ```
    *   `DELETE /responsibility-groups/{groupId}` (取消用户职责分配)
        *   **Response (Success - 204 No Content)**: (无响应体)

*   **环境管理** (`/api/v1/environments`):
    *   `GET /`
        *   **Query Params**: `page=1`, `pageSize=10`, `status=active`, `type=...`, `code=...`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, "message": "Success", "data": { "items": [ { "id": 1, "name": "Production", "code": "prod", "description": "...", "type": "cloud", "status": "active", "createdAt": "...", "updatedAt": "..." } ], "total": ..., "page": ..., "pageSize": ... } }`
    *   `POST /`
        *   **Request Body**: `{ "name": "Staging", "code": "staging", "description": "Pre-release testing", "type": "hybrid", "status": "active" }`
        *   **Response (Success - 201 Created)**: `{ "code": 0, "message": "Environment created", "data": { "id": 2, ... } }`
    *   `GET /{envId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "name": "Production",
                "code": "prod",
                "description": "...",
                "type": "cloud",
                "status": "active",
                "createdAt": "...",
                "updatedAt": "...",
                "serviceInstanceCount": 15,
                "openBugCount": 5
              }
            }
            ```
    *   `PUT /{envId}`
        *   **Request Body**: `{ "description": "Updated description", "status": "maintenance" }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Environment updated", "data": { "id": 1, ... } }`
    *   `DELETE /{envId}`
        *   **Response (Success - 204 No Content)**

*   **资产管理 (服务器)** (`/api/v1/assets`):
    *   `GET /?type=server`
        *   **Query Params**: `page`, `pageSize`, `status`, `ipAddress`, `hostname`, `os`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, "message": "Success", "data": { "items": [ { "id": 1, "assetId": 101, "name": "Web Server 01", "type": "server", "status": "in_use", "ipAddress": "192.168.1.10", "hostname": "web01", "os": "Ubuntu 22.04", ... } ], "total": ..., "page": ..., "pageSize": ... } }` (assetId 对应 assets 表 ID，其他为 server_assets 字段)
    *   `POST /`
        *   **Request Body**: `{ "name": "DB Server 01", "type": "server", "status": "in_use", "serverDetails": { "ipAddress": "192.168.1.20", "hostname": "db01", "os": "CentOS 8", ... } }`
        *   **Response (Success - 201 Created)**: `{ "code": 0, "message": "Asset created", "data": { "id": 2, "assetId": 102, ... } }`
    *   `GET /{assetId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "assetId": 101,
                "name": "Web Server 01",
                "type": "server",
                "status": "in_use",
                "ipAddress": "192.168.1.10",
                "hostname": "web01",
                "os": "Ubuntu 22.04",
                "cpuCores": 8,
                "memoryGb": 32,
                "diskGb": 500,
                "location": "Datacenter A, Rack 5",
                "createdAt": "...",
                "updatedAt": "...",
                "runningInstances": [
                  {
                    "instanceId": 5,
                    "service": { "id": 1, "name": "User Service" },
                    "port": 8080,
                    "version": "v1.2.0",
                    "status": "running"
                  },
                  {
                    "instanceId": 9,
                    "service": { "id": 2, "name": "Order Service" },
                    "port": 8081,
                    "version": "v1.1.0",
                    "status": "running"
                  }
                ]
              }
            }
            ```
    *   `PUT /{assetId}`
        *   **Request Body**: `{ "name": "Web Server 01 (Updated)", "status": "maintenance", "serverDetails": { "os": "Ubuntu 24.04" } }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, "message": "Asset updated", "data": { "id": 1, "assetId": 101, ... } }`
    *   `DELETE /{assetId}`
        *   **Response (Success - 204 No Content)**

*   **服务管理** (`/api/v1/services`):
    *   `GET /`
        *   **Query Params**: `page`, `pageSize`, `name`, `serviceTypeId`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, ..., "data": { "items": [ { "id": 1, "name": "User Service", "description": "Handles user auth and profile", "serviceType": { "id": 1, "name": "API" }, "createdAt": "..." } ], ... } }`
    *   `POST /`
        *   **Request Body**: `{ "name": "Order Service", "description": "...", "serviceTypeId": 1 }`
        *   **Response (Success - 201 Created)**: `{ "code": 0, ..., "data": { "id": 2, ... } }`
    *   `GET /{serviceId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "name": "User Service",
                "description": "Handles user auth and profile",
                "serviceType": {
                  "id": 1,
                  "name": "API"
                },
                "instanceCount": 8,
                "openBugCount": 3,
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
    *   `PUT /{serviceId}`
        *   **Request Body**: `{ "description": "Updated description" }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, ..., "data": { "id": 1, ... } }`
    *   `DELETE /{serviceId}`
        *   **Response (Success - 204 No Content)**
    *   `GET /types` (获取所有服务类型)
        *   **Response (Success - 200 OK)**: `{ "code": 0, ..., "data": [ { "id": 1, "name": "API", "description": "..." } ] }`

*   **服务实例管理** (`/api/v1/service-instances`):
    *   `GET /`
        *   **Query Params**: `page`, `pageSize`, `environmentId`, `serviceId`, `serverAssetId`, `status`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, ..., "data": { "items": [ { "id": 1, "service": { "id": 1, "name": "User Service" }, "environment": { "id": 1, "code": "prod" }, "server": { "assetId": 101, "hostname": "web01" }, "port": 8080, "status": "running", "version": "v1.2.0", "createdAt": "..." } ], ... } }`
    *   `POST /`
        *   **Request Body**: `{ "serviceId": 1, "environmentId": 2, "serverAssetId": 103, "port": 8081, "status": "stopped", "version": "v1.3.0" }`
        *   **Response (Success - 201 Created)**: `{ "code": 0, ..., "data": { "id": 2, ... } }`
    *   `GET /{instanceId}`
        *   **Response (Success - 200 OK)**: `{ "code": 0, ..., "data": { "id": 1, ... } }`
    *   `PUT /{instanceId}`
        *   **Request Body**: `{ "status": "running", "version": "v1.3.1" }`
        *   **Response (Success - 200 OK)**: `{ "code": 0, ..., "data": { "id": 1, ... } }`
    *   `DELETE /{instanceId}`
        *   **Response (Success - 204 No Content)**

*   **业务管理** (`/api/v1/businesses`):
    *   `GET /`
        *   **Query Params**: `page=1`, `pageSize=20`, `name=...`, `code=...`, `status=active`, `ownerUserId=...`
        *   **Response (Success - 200 OK)**: (分页结构)
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "items": [
                  {
                    "id": 1,
                    "name": "E-commerce Platform",
                    "code": "ECOMM",
                    "description": "Main online retail platform",
                    "owner": {
                       "id": 15,
                       "name": "Product Manager A"
                    },
                    "status": "active",
                    "createdAt": "...",
                    "updatedAt": "..."
                  },
                  {
                    "id": 2,
                    "name": "Internal CRM",
                    "code": "CRM",
                    "description": "Customer Relationship Management system",
                     "owner": {
                       "id": 21,
                       "name": "Sales Lead"
                    },
                    "status": "active",
                    "createdAt": "...",
                    "updatedAt": "..."
                  }
                ],
                "total": 10,
                "page": 1,
                "pageSize": 20
              }
            }
            ```
    *   `POST /`
        *   **Request Body**:
            ```json
            {
              "name": "New Mobile App Project",
              "code": "MOBILE_APP",
              "description": "Development of the new native mobile application",
              "ownerUserId": 30,
              "status": "active"
            }
            ```
        *   **Response (Success - 201 Created)**:
            ```json
            {
              "code": 0,
              "message": "Business created successfully",
              "data": {
                "id": 3,
                "name": "New Mobile App Project",
                "code": "MOBILE_APP",
                "description": "Development of the new native mobile application",
                "ownerUserId": 30,
                "status": "active",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40007, "message": "Validation error: Business name or code already exists", "data": null }`
    *   `GET /{businessId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "name": "E-commerce Platform",
                "code": "ECOMM",
                "description": "Main online retail platform",
                "owner": {
                   "id": 15,
                   "name": "Product Manager A"
                },
                "status": "active",
                "linkedServiceCount": 5,
                "openBugCount": 12,
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
        *   **Response (Error - 404 Not Found)**: `{ "code": 40404, "message": "Business not found", "data": null }`
    *   `PUT /{businessId}`
        *   **Request Body**:
            ```json
            {
              "description": "Main online retail platform - International Expansion Phase",
              "ownerUserId": 35,
              "status": "active"
            }
            ```
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Business updated successfully",
              "data": {
                "id": 1,
                "name": "E-commerce Platform",
                "code": "ECOMM",
                "description": "Main online retail platform - International Expansion Phase",
                "ownerUserId": 35,
                "status": "active",
                "createdAt": "...",
                "updatedAt": "..."
              }
            }
            ```
    *   `DELETE /{businessId}`
        *   **Response (Success - 204 No Content)**: (无响应体)
        *   **Response (Error - 400 Bad Request)**: `{ "code": 40008, "message": "Cannot delete business: It has linked resources (e.g., bugs, services)", "data": null }` (需要检查关联)

*   **Bug 管理** (`/api/v1/bugs`):
    *   `GET /`
        *   **Query Params**: `page`, `pageSize`, `status`, `priority`, `reporterId`, `assigneeGroupId`, `environmentId`, `serviceInstanceId`, `businessId`, `sortBy=createdAt`, `order=desc`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, ..., "data": { "items": [ { "id": 1, "title": "Login button not working", "status": "open", "priority": "high", "reporter": { "id": 5, "name": "Tester" }, "assigneeGroup": { "id": 10, "responsibilityName": "UI Dev Team" }, "environment": { "id": 2, "code": "staging" }, "createdAt": "..." } ], ... } }`
    *   `POST /`
        *   **Request Body**: `{ "title": "...", "description": "...", "priority": "medium", "environmentId": 2, "serviceInstanceId": 5, "businessId": 3, "reporterId": 5 }` (assigneeGroupId 可由后端根据业务/服务自动分配或后续手动指定)
        *   **Response (Success - 201 Created)**: `{ "code": 0, ..., "data": { "id": 2, ... } }`
    *   `GET /{bugId}`
        *   **Response (Success - 200 OK)**:
            ```json
            {
              "code": 0,
              "message": "Success",
              "data": {
                "id": 1,
                "title": "Login button not working",
                "description": "Steps to reproduce... Expected result... Actual result...",
                "status": "open",
                "priority": "high",
                "reporter": {
                  "id": 5,
                  "name": "Tester",
                  "email": "tester@example.com"
                },
                "assigneeGroup": {
                  "id": 10,
                  "responsibilityName": "UI Dev Team"
                },
                "environment": {
                  "id": 2,
                  "name": "Staging",
                  "code": "staging"
                },
                "serviceInstance": {
                  "id": 5,
                  "serviceName": "Frontend Gateway",
                  "version": "v0.8.1",
                  "serverHostname": "web02"
                },
                "business": {
                  "id": 1,
                  "name": "E-commerce Platform",
                  "code": "ECOMM"
                },
                "createdAt": "...",
                "updatedAt": "...",
                "resolvedAt": null,
                "closedAt": null,
                "comments": [
                  {
                    "id": 1,
                    "author": { "id": 10, "name": "Dev Lead" },
                    "text": "Investigating...",
                    "createdAt": "..."
                  }
                ],
                "attachments": [
                  {
                    "id": 1,
                    "filename": "screenshot.png",
                    "url": "/api/v1/attachments/1/download",
                    "uploadedBy": { "id": 5, "name": "Tester" },
                    "createdAt": "..."
                  }
                ],
                "history": [
                   { "timestamp": "...", "user": { "id": 1, "name": "Admin"}, "action": "status_change", "from": "open", "to": "in_progress"},
                   { "timestamp": "...", "user": { "id": 1, "name": "Admin"}, "action": "assignee_change", "to": { "groupId": 11, "name": "Backend Team"}}
                ]
              }
            }
            ```
    *   `PUT /{bugId}`
        *   **Request Body**: `{ "status": "in_progress", "assigneeGroupId": 11 }` (或 `{ "comment": "Starting investigation" }`)
        *   **Response (Success - 200 OK)**: `{ "code": 0, ..., "data": { "id": 1, ... } }`
    *   (可能需要 `PATCH` 来部分更新，如只添加评论)

*   **审计日志查询** (`/api/v1/audit-logs`):
    *   `GET /`
        *   **Query Params**: `page`, `pageSize`, `userId`, `action`, `targetType`, `targetId`, `startDate`, `endDate`
        *   **Response (Success - 200 OK)**: (分页结构) `{ "code": 0, ..., "data": { "items": [ { "id": 1, "user": { "id": 1, "name": "Admin" }, "timestamp": "...", "action": "update", "targetType": "environment", "targetId": 1, "details": "Status changed from active to maintenance" } ], ... } }`

## 5. 分页、排序与过滤

*   对于返回列表的 GET 请求 (如 `/users`, `/environments`, `/bugs`)，应支持分页参数 (如 `page`, `pageSize`)。
*   支持通过查询参数进行排序 (如 `sortBy=createdAt&order=desc`)。
*   支持通过查询参数进行过滤 (如 `GET /bugs?status=open&assigneeId=123`)。
