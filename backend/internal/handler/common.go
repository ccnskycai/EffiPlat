package handler

import (
	"EffiPlat/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// RespondWithError sends a JSON error response.
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, model.ErrorResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// RespondWithSuccess sends a JSON success response.
func RespondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = "success"
	}
	c.JSON(code, model.SuccessResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}
