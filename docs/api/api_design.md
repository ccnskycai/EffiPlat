 # API 设计规范

## 1. 设计原则

*   **风格**: 遵循 RESTful 设计风格。
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
    {
      "code": 0, // 0 表示成功，非 0 表示错误
      "message": "Success",
      "data": { ... } // 或 [...] 或 null
    }
    // 错误示例
    {
      "code": 4001, // 自定义业务错误码
      "message": "Invalid input parameter: email format error",
      "data": null
    }
    ```

## 2. 认证与授权

*   **认证**: 初步考虑使用基于 Token 的认证方式 (如 JWT)。用户登录成功后，服务端颁发 Token，客户端后续请求在 Header 中携带 Token (`Authorization: Bearer <token>`)。
*   **授权**: 基于 RBAC 模型。服务端在处理请求时，根据用户角色和预定义的权限规则，判断用户是否有权执行该操作。

## 3. 版本管理

*   API 应考虑版本管理，以支持未来的非兼容性变更。初步考虑在 URL 中加入版本号 (如 `/api/v1/...`)。

## 4. 核心 API 端点 (V1.0 关注点)

[根据 V1.0 计划 (`docs/requirements/execution_plan.md`)，定义核心资源的操作接口]

*   **用户认证**:
    *   `POST /api/v1/auth/login`
    *   `POST /api/v1/auth/logout`
*   **用户管理**:
    *   `GET /api/v1/users`
    *   `POST /api/v1/users`
    *   `GET /api/v1/users/{userId}`
    *   `PUT /api/v1/users/{userId}`
*   **角色与权限管理** (管理员接口):
    *   `GET /api/v1/roles`
    *   `POST /api/v1/roles`
    *   ...
*   **职责与职责组管理**:
    *   `GET /api/v1/responsibilities`
    *   `POST /api/v1/responsibilities`
    *   `GET /api/v1/responsibility-groups`
    *   `POST /api/v1/responsibility-groups`
    *   ...
*   **环境管理**:
    *   `GET /api/v1/environments`
    *   `POST /api/v1/environments`
    *   `GET /api/v1/environments/{envId}`
    *   `PUT /api/v1/environments/{envId}`
*   **资产管理 (服务器)**:
    *   `GET /api/v1/assets?type=server`
    *   `POST /api/v1/assets` (需指定类型为 server)
    *   `GET /api/v1/assets/{assetId}`
    *   `PUT /api/v1/assets/{assetId}`
*   **服务管理**:
    *   `GET /api/v1/services`
    *   `POST /api/v1/services`
    *   `GET /api/v1/services/{serviceId}`
    *   `PUT /api/v1/services/{serviceId}`
*   **服务实例管理**:
    *   `GET /api/v1/service-instances?environmentId={envId}&serviceId={serviceId}`
    *   `POST /api/v1/service-instances`
    *   ...
*   **业务管理**:
    *   `GET /api/v1/businesses`
    *   `POST /api/v1/businesses`
    *   ...
*   **Bug 管理**:
    *   `GET /api/v1/bugs` (支持按状态、负责人、环境等过滤)
    *   `POST /api/v1/bugs`
    *   `GET /api/v1/bugs/{bugId}`
    *   `PUT /api/v1/bugs/{bugId}` (更新状态、负责人等)
*   **审计日志查询**:
    *   `GET /api/v1/audit-logs` (支持按用户、时间、操作对象过滤)

[待补充: 每个核心接口的详细请求参数、请求体结构、响应体结构]

## 5. 分页、排序与过滤

*   对于返回列表的 GET 请求 (如 `/users`, `/environments`, `/bugs`)，应支持分页参数 (如 `page`, `pageSize`)。
*   支持通过查询参数进行排序 (如 `sortBy=createdAt&order=desc`)。
*   支持通过查询参数进行过滤 (如 `GET /bugs?status=open&assigneeId=123`)。
