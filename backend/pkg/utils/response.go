package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse is a consistent structure for API responses.
type StandardResponse struct {
	Code    int         `json:"code"`    // Business-specific code, 0 for success
	Message string      `json:"message"` // Descriptive message
	Data    interface{} `json:"data,omitempty"` // Payload, omitted if nil
}

// SendSuccessResponse sends a structured success response.
func SendSuccessResponse(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, StandardResponse{
		Code:    0, // Typically 0 for success
		Message: message,
		Data:    data,
	})
}

// SendErrorResponse sends a structured error response.
// errorCode is a business-specific error code.
func SendErrorResponse(c *gin.Context, httpStatus int, message string, details ...string) {
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