# API 设计规范

## 1. 设计原则

- **风格**: 遵循 RESTful 设计风格。
- **传输安全**: 所有 API 请求必须通过 HTTPS 进行以确保传输安全。
- **数据格式**: 请求体和响应体统一使用 JSON 格式。
- **命名**: URL 路径使用小写字母和连字符 (`-`)，资源名使用复数形式 (如 `/users`, `/environments`)。
- **HTTP 方法**: 合理使用 HTTP 方法表达操作意图：
  - `GET`: 获取资源。
  - `POST`: 创建资源。
  - `PUT`: 完整更新资源。
  - `PATCH`: 部分更新资源。
  - `DELETE`: 删除资源。
- **状态码**: 使用标准的 HTTP 状态码表示请求结果：
  - `200 OK`: 请求成功 (GET, PUT, PATCH)。
  - `201 Created`: 资源创建成功 (POST)。
  - `204 No Content`: 请求成功，但无返回内容 (DELETE)。
  - `400 Bad Request`: 客户端请求错误（如参数校验失败）。
  - `401 Unauthorized`: 未认证。
  - `403 Forbidden`: 已认证但无权限访问。
  - `404 Not Found`: 请求的资源不存在。
  - `500 Internal Server Error`: 服务端内部错误。
- **统一响应格式**: 设计统一的 JSON 响应结构，包含状态码/业务码、消息和数据。
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

- **认证**: 初步考虑使用基于 Token 的认证方式 (如 JWT)。用户登录成功后，服务端颁发 Token，客户端后续请求在 Header 中携带 Token (`Authorization: Bearer <token>`)。
- **授权**: 基于 RBAC 模型。服务端在处理请求时，根据用户角色和预定义的权限规则，判断用户是否有权执行该操作。

## 3. 版本管理

- API 应考虑版本管理，以支持未来的非兼容性变更。初步考虑在 URL 中加入版本号 (如 `/api/v1/...`)。

## 4. 核心 API 端点 (V1.0 关注点)

根据 V1.0 计划 (`docs/requirements/execution_plan.md`)，定义核心资源的操作接口

## 5. 分页、排序与过滤

- 对于返回列表的 GET 请求 (如 `/users`, `/environments`, `/bugs`)，应支持分页参数 (如 `page`, `pageSize`)。
- 支持通过查询参数进行排序 (如 `sortBy=createdAt&order=desc`)。
- 支持通过查询参数进行过滤 (如 `GET /bugs?status=open&assigneeId=123`)。

[EffiPlat OpenAPI 3.0 文档（YAML）](./api/openapi.yaml)
