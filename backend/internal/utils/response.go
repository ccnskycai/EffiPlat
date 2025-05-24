package utils

import (
	"net/http"

	"EffiPlat/backend/internal/models" // Import models for response structs

	"github.com/gin-gonic/gin"
)

// SendSuccessResponse sends a standardized success JSON response.
// It uses the models.SuccessResponse structure.
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

	c.JSON(httpStatusCode, models.SuccessResponse{
		Code:    0, // Typically 0 for business logic success
		Message: message,
		Data:    data,
	})
}

// SendErrorResponse sends a standardized error JSON response.
// It uses the models.ErrorResponse structure.
// The 'details' parameter is variadic, but we'll only use the first element if provided.
func SendErrorResponse(c *gin.Context, httpStatusCode int, message string, details ...string) {
	// Note: models.ErrorResponse has a 'Data interface{}' field which can be used for 'details'.
	// For simplicity, we are putting string details into the Message field if only one is provided after the main message,
	// or using the 'Data' field if a more structured error detail is needed (not implemented here).
	// For now, this implementation will just use the message. The 'details' are ignored to match current ErrorResponse struct.
	// If `details` were to be used, models.ErrorResponse would need a `Details string` field or similar.
	// Let's assume the current models.ErrorResponse's `Data` field can be used for simple string details.

	var errorData interface{}
	if len(details) > 0 {
		errorData = details[0] // Use the first detail string for the Data field
	}

	c.JSON(httpStatusCode, models.ErrorResponse{
		Code:    httpStatusCode, // Using HTTP status as the primary error code
		Message: message,
		Data:    errorData, // Populate Data field with details if any
	})
}

// SendPaginatedSuccessResponse sends a standardized success JSON response for paginated data.
// It wraps the paginated data within the standard SuccessResponse using models.PaginatedData.
func SendPaginatedSuccessResponse(c *gin.Context, httpStatusCode int, message string, items interface{}, page int, pageSize int, total int64) {
	paginatedData := models.PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	c.JSON(httpStatusCode, models.SuccessResponse{
		Code:    0,
		Message: message,
		Data:    paginatedData,
	})
}
