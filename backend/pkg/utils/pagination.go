package utils

// PaginatedResult defines the structure for paginated API responses.
// It uses a generic type T for the items, making it reusable.
type PaginatedResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
} 