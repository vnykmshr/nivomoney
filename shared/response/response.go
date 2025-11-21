// Package response provides standardized HTTP response formats for Nivo APIs.
package response

import (
	"encoding/json"
	"net/http"

	"github.com/vnykmshr/nivo/shared/errors"
)

// Response represents a standardized API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorData contains error information.
type ErrorData struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Meta contains metadata about the response.
type Meta struct {
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// JSON writes a JSON response with the given status code and data.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// Success writes a success response.
func Success(w http.ResponseWriter, statusCode int, data interface{}) error {
	return JSON(w, statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta writes a success response with metadata.
func SuccessWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *Meta) error {
	return JSON(w, statusCode, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error writes an error response from an errors.Error.
func Error(w http.ResponseWriter, err *errors.Error) error {
	statusCode := err.HTTPStatusCode()

	return JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorData{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
	})
}

// ErrorWithMeta writes an error response with metadata.
func ErrorWithMeta(w http.ResponseWriter, err *errors.Error, meta *Meta) error {
	statusCode := err.HTTPStatusCode()

	return JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorData{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
		Meta: meta,
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) error {
	return Error(w, errors.BadRequest(message))
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) error {
	return Error(w, errors.Unauthorized(message))
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) error {
	return Error(w, errors.Forbidden(message))
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, resource string) error {
	return Error(w, errors.NotFound(resource))
}

// Conflict writes a 409 Conflict response.
func Conflict(w http.ResponseWriter, message string) error {
	return Error(w, errors.Conflict(message))
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) error {
	return Error(w, errors.Internal(message))
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) error {
	return Success(w, http.StatusCreated, data)
}

// OK writes a 200 OK response.
func OK(w http.ResponseWriter, data interface{}) error {
	return Success(w, http.StatusOK, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Paginated writes a paginated response with metadata.
func Paginated(w http.ResponseWriter, data interface{}, page, pageSize int, totalItems int64) error {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return SuccessWithMeta(w, http.StatusOK, data, &Meta{
		Pagination: pagination,
	})
}

// ValidationError writes a validation error response.
func ValidationError(w http.ResponseWriter, details map[string]interface{}) error {
	err := errors.Validation("validation failed")
	err.Details = details
	return Error(w, err)
}
