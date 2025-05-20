package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UnifiedResponse is a standard structure for API responses.
// Use this for consistency across the API.
type UnifiedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedData is a structure for paginated list responses.
// It's typically used within the Data field of UnifiedResponse.
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

func Respond(c *gin.Context, httpStatus int, code int, message string, data interface{}) {
	c.JSON(httpStatus, UnifiedResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func OK(c *gin.Context, data interface{}) {
	Respond(c, http.StatusOK, 0, "Success", data)
}

func Created(c *gin.Context, data interface{}) {
	Respond(c, http.StatusCreated, 0, "Resource created successfully", data)
}

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

func Paginated(c *gin.Context, items interface{}, total int64, page int, pageSize int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10 // Default, should ideally match handler's default
	}
	totalPages := 0
	if total > 0 && pageSize > 0 {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}

	Respond(c, http.StatusOK, 0, "Success", PaginatedData{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func Error(c *gin.Context, httpStatus int, message string) {
	Respond(c, httpStatus, httpStatus, message, nil)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}
