package handler

import (
	"EffiPlat/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// RespondWithError sends a JSON error response.
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, models.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// RespondWithSuccess sends a JSON success response.
func RespondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = "success"
	}
	c.JSON(code, models.SuccessResponse{
		BizCode: 0,
		Message: message,
		Data:    data,
	})
}
