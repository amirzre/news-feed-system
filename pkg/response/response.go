package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Message string     `json:"message,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse wraps paginated data
type PaginatedResponse struct {
	Items      any               `json:"items"`
	Pagination *PaginationInfo   `json:"pagination"`
	Filters    map[string]string `json:"filters,omitempty"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Success returns a successful response
func Success(c echo.Context, statusCode int, data any, message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	response := APIResponse{
		Success: true,
		Data:    data,
		Message: msg,
	}

	return c.JSON(statusCode, response)
}

// SuccessWithPagination returns a successful response with pagination
func SuccessWithPagination(c echo.Context, items any, pagination *PaginationInfo, filters map[string]string, message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	data := PaginatedResponse{
		Items:      items,
		Pagination: pagination,
		Filters:    filters,
	}

	response := APIResponse{
		Success: true,
		Data:    data,
		Message: msg,
	}

	return c.JSON(http.StatusOK, response)
}

// Error returns an error response
func Error(c echo.Context, statusCode int, message string, details ...string) error {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	errorInfo := &ErrorInfo{
		Message: message,
		Details: detail,
	}

	response := APIResponse{
		Success: false,
		Error:   errorInfo,
	}

	return c.JSON(statusCode, response)
}

// InternalServerError returns a 500 error response
func InternalServerError(c echo.Context, message string, details ...string) error {
	return Error(c, http.StatusInternalServerError, message, details...)
}

// BadRequest returns a 400 error response
func BadRequest(c echo.Context, message string, details ...string) error {
	return Error(c, http.StatusBadRequest, message, details...)
}

// NotFound returns a 404 error response
func NotFound(c echo.Context, message string, details ...string) error {
	return Error(c, http.StatusNotFound, message, details...)
}

// Conflict returns a 409 error response
func Conflict(c echo.Context, message string, details ...string) error {
	return Error(c, http.StatusConflict, message, details...)
}

func ValidationError(c echo.Context, err error) error {
	errorInfo := &ErrorInfo{
		Message: "Request validation failed",
		Details: err.Error(),
	}

	response := APIResponse{
		Success: false,
		Error:   errorInfo,
	}

	return c.JSON(http.StatusBadRequest, response)
}

// CreatePaginationInfo creates pagination metadata
func CreatePaginationInfo(page, limit, total int) *PaginationInfo {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
