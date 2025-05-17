package models

// SuccessResponse defines the structure for a successful API response.
type SuccessResponse struct {
	BizCode int         `json:"bizCode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse defines the structure for an error API response.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PaginatedData defines the structure for paginated data within a SuccessResponse.
type PaginatedData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}
