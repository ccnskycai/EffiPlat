package utils

import (
	"net/http"

	"EffiPlat/backend/internal/model" // Import models for response structs

	"github.com/gin-gonic/gin"
)

// StandardResponse is a consistent structure for API responses.
type StandardResponse struct {
	Code    int         `json:"code"`    // Business-specific code, 0 for success
	Message string      `json:"message"` // Descriptive message
	Data    interface{} `json:"data,omitempty"` // Payload, omitted if nil
}

// SendStandardSuccessResponse sends a structured success response using StandardResponse structure.
func SendStandardSuccessResponse(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, StandardResponse{
		Code:    0, // Typically 0 for success
		Message: message,
		Data:    data,
	})
}

// SendStandardErrorResponse sends a structured error response using StandardResponse structure.
// errorCode is a business-specific error code.
func SendStandardErrorResponse(c *gin.Context, httpStatus int, message string, details ...string) {
	// Use the first detail as the data payload if present, otherwise keep it simple
	var errorData interface{}
	if len(details) > 0 {
		errorData = details[0] // Or can be a map[string]string{"details": details[0]}
	}

	// Infer a business error code from httpStatus if not directly provided or make it generic
	// For simplicity, let's use a convention like httpStatus * 100 + 1 for now
	// Example: 400 Bad Request -> 40001
	// This needs to align with your API design doc's error code strategy.
	businessErrorCode := httpStatus * 100 // Simplified, adjust as per your error code design
	if httpStatus == http.StatusNotFound { // Example specific code
		businessErrorCode = 40401
	} else if httpStatus == http.StatusBadRequest { // Example specific code
		businessErrorCode = 40001
	} else if httpStatus == http.StatusForbidden { // Example specific code
		businessErrorCode = 40301
	}

	c.JSON(httpStatus, StandardResponse{
		Code:    businessErrorCode,
		Message: message,
		Data:    errorData,
	})
}

// SendSuccessResponse sends a standardized success JSON response.
// It uses the model.SuccessResponse structure.
func SendSuccessResponse(c *gin.Context, httpStatusCode int, data interface{}, customMessage ...string) {
	// Handle cases like 204 No Content where no body should be sent
	if data == nil && (httpStatusCode == http.StatusNoContent || httpStatusCode == http.StatusAccepted) {
		c.Status(httpStatusCode)
		return
	}

	message := ""
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	} else {
		switch httpStatusCode {
		case http.StatusCreated:
			message = "Resource created successfully"
		case http.StatusAccepted:
			message = "Request accepted"
		case http.StatusNoContent:
			message = "Operation successful, no content to return"
		default:
			message = "Success"
		}
	}

	c.JSON(httpStatusCode, model.SuccessResponse{
		Code:    0, // Typically 0 for business logic success
		Message: message,
		Data:    data,
	})
}

// SendErrorResponse sends a standardized error JSON response.
// It uses the model.ErrorResponse structure.
// The 'details' parameter is variadic, but we'll only use the first element if provided.
func SendErrorResponse(c *gin.Context, httpStatusCode int, message string, details ...string) {
	// Note: model.ErrorResponse has a 'Data interface{}' field which can be used for 'details'.
	// For simplicity, we are putting string details into the Message field if only one is provided after the main message,
	// or using the 'Data' field if a more structured error detail is needed (not implemented here).
	// For now, this implementation will just use the message. The 'details' are ignored to match current ErrorResponse struct.
	// If `details` were to be used, model.ErrorResponse would need a `Details string` field or similar.
	// Let's assume the current model.ErrorResponse's `Data` field can be used for simple string details.

	var errorData interface{}
	if len(details) > 0 {
		errorData = details[0] // Use the first detail string for the Data field
	}

	c.JSON(httpStatusCode, model.ErrorResponse{
		Code:    httpStatusCode, // Using HTTP status as the primary error code
		Message: message,
		Data:    errorData, // Populate Data field with details if any
	})
}

// SendPaginatedSuccessResponse sends a standardized success JSON response for paginated data.
// It wraps the paginated data within the standard SuccessResponse using model.PaginatedData.
func SendPaginatedSuccessResponse(c *gin.Context, httpStatusCode int, message string, items interface{}, page int, pageSize int, total int64) {
	paginatedData := model.PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	c.JSON(httpStatusCode, model.SuccessResponse{
		Code:    0,
		Message: message,
		Data:    paginatedData,
	})
}

// PaginatedResponse defines the structure for paginated API responses.
// It uses interface{} for Items to be more generic, as specific types are in model.BugResponse etc.
// The handlers will pass the correctly typed slice to NewPaginatedResponse.
// The swagger doc in handler will specify the actual data type e.g. []model.BugResponse
type PaginatedResponse struct {
	Items      interface{} `json:"items"`     // 统一使用Items字段名
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// NewPaginatedResponse creates a new PaginatedResponse.
func NewPaginatedResponse(data interface{}, totalCount int64, page int, pageSize int) PaginatedResponse {
	totalPages := 0
	if pageSize > 0 && totalCount > 0 {
		totalPages = (int(totalCount) + pageSize - 1) / pageSize
	} else if totalCount == 0 {
		totalPages = 0 // No pages if no items
	} else if pageSize == 0 && totalCount > 0 {
		totalPages = 1 // Assume all items on one page if pageSize is 0 but items exist
	}

	return PaginatedResponse{
		Items:      data,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// 以下是从pkg/response添加的函数，用于兼容现有代码

// Respond 发送标准化的JSON响应
func Respond(c *gin.Context, httpStatus int, code int, message string, data interface{}) {
	c.JSON(httpStatus, model.SuccessResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// OK 发送200 OK的成功响应
func OK(c *gin.Context, data interface{}) {
	Respond(c, http.StatusOK, 0, "Success", data)
}

// Created 发送201 Created的成功响应
func Created(c *gin.Context, data interface{}) {
	Respond(c, http.StatusCreated, 0, "Resource created successfully", data)
}

// Status 发送指定HTTP状态码的响应，无数据
func Status(c *gin.Context, httpStatus int) {
	if httpStatus == http.StatusNoContent {
		c.Status(http.StatusNoContent)
		return
	}
	message := http.StatusText(httpStatus)
	if message == "" {
		message = "Request processed successfully"
	}
	Respond(c, httpStatus, 0, message, nil)
}

// Paginated 发送分页数据的成功响应
func Paginated(c *gin.Context, items interface{}, total int64, page int, pageSize int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10 // 默认值，应与handler的默认值匹配
	}
	// 计算总页数但不需要传递给现有的PaginatedData结构
	// 这是为了与现有结构兼容
	//totalPages := 0
	//if total > 0 && pageSize > 0 {
	//	totalPages = (int(total) + pageSize - 1) / pageSize
	//}

	paginatedData := model.PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	Respond(c, http.StatusOK, 0, "Success", paginatedData)
}

// Error 发送错误响应
func Error(c *gin.Context, httpStatus int, message string) {
	Respond(c, httpStatus, httpStatus, message, nil)
}

// BadRequest 发送400错误响应
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound 发送404错误响应
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 发送500错误响应
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Unauthorized 发送401错误响应
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 发送403错误响应
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}
