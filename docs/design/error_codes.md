# 错误代码规范

本文档定义了"EffiPlat"后端服务使用的内部错误代码规范。

## 1. 目的

*   提供一套标准化的错误代码，用于内部日志记录、监控告警和问题追踪。
*   作为将内部错误映射到对外 API 错误响应的基础。
*   方便开发人员理解和处理特定错误场景。

## 2. 格式规范

*   错误代码建议采用**字符串**格式，使用**大写字母和下划线** (`UPPER_SNAKE_CASE`)，以提高可读性。
*   错误代码应具有一定的层级或分类含义，例如按模块或错误类型划分。
*   格式示例：`[模块/分类]_[具体错误]` (例如 `VALIDATION_INVALID_INPUT`, `DATABASE_RECORD_NOT_FOUND`)

## 3. 错误代码定义 (初步)

| 错误代码 (Code)               | HTTP 状态码 (建议) | 描述 (中文)                     | 描述 (英文)                           |
| :---------------------------- | :----------------- | :------------------------------ | :------------------------------------ |
| **通用错误 (COMMON)**         |                    |                                 |                                       |
| `UNKNOWN_ERROR`               | 500                | 未知错误/服务器内部错误         | Unknown internal server error         |
| `SERVICE_UNAVAILABLE`         | 503                | 服务暂时不可用                   | Service temporarily unavailable       |
| **认证与授权 (AUTH)**     |                    |                                 |                                       |
| `AUTH_INVALID_CREDENTIALS`    | 401                | 无效的凭证 (用户名或密码错误) | Invalid credentials                   |
| `AUTH_INVALID_TOKEN`          | 401                | 无效或过期的 Token             | Invalid or expired token              |
| `AUTH_PERMISSION_DENIED`      | 403                | 权限不足                         | Permission denied                     |
| **输入验证 (VALIDATION)**     |                    |                                 |                                       |
| `VALIDATION_INVALID_INPUT`    | 400                | 无效的输入参数                   | Invalid input parameters              |
| `VALIDATION_MISSING_FIELD`    | 400                | 缺少必要的字段                   | Required field is missing             |
| `VALIDATION_INVALID_FORMAT`   | 400                | 字段格式错误                     | Invalid field format                  |
| `VALIDATION_VALUE_OUT_OF_RANGE`| 400                | 字段值超出范围                   | Field value out of range              |
| **资源相关 (RESOURCE)**       |                    |                                 |                                       |
| `RESOURCE_NOT_FOUND`          | 404                | 请求的资源未找到                 | Requested resource not found          |
| `RESOURCE_ALREADY_EXISTS`     | 409                | 尝试创建已存在的资源             | Resource already exists               |
| **数据库操作 (DATABASE)**     |                    |                                 |                                       |
| `DATABASE_CONNECTION_ERROR`   | 500                | 数据库连接错误                   | Database connection error             |
| `DATABASE_QUERY_ERROR`        | 500                | 数据库查询错误                   | Database query error                  |
| `DATABASE_RECORD_NOT_FOUND`   | 404                | 数据库记录未找到 (内部使用)     | Database record not found             |
| `DATABASE_DUPLICATE_ENTRY`    | 409                | 数据库记录重复 (唯一约束冲突) | Database duplicate entry              |
| **外部服务 (EXTERNAL)**       |                    |                                 |                                       |
| `EXTERNAL_SERVICE_TIMEOUT`    | 504                | 调用外部服务超时                 | External service timeout              |
| `EXTERNAL_SERVICE_ERROR`      | 502                | 调用外部服务失败                 | External service error                |

**注意:**

*   `DATABASE_RECORD_NOT_FOUND` 通常在 Service 层被转换为更通用的 `RESOURCE_NOT_FOUND` 返回给客户端。
*   此列表为初步定义，随着开发进展会不断补充和细化。
*   每个错误代码应在代码中有明确的定义和使用。

## 4. 在代码中使用

建议定义常量或枚举来表示这些错误代码，并在自定义错误类型中引用它们。

```go
// 示例
const (
	ErrCodeValidationInvalidInput = "VALIDATION_INVALID_INPUT"
	ErrCodeResourceNotFound       = "RESOURCE_NOT_FOUND"
)

type ErrValidation struct {
	Code        string
	FieldErrors map[string]string
}

func (e *ErrValidation) Error() string { return e.Code } // Error() 可以只返回 Code
``` 