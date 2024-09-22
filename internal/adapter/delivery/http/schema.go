package http

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

const statusError = "error"

// urlRequest represents the structure for a request to shorten or modifying a URL.
type urlRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}

// urlResponse represents the structure for a response containing shortened URL information.
type urlResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toURLResponse converts an entity.URL to a urlResponse.
func toURLResponse(url *entity.URL) urlResponse {
	return urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}
}

// urlStatsResponse represents the structure for a response containing URL statistics.
type urlStatsResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Stats       urlStats  `json:"stats"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// urlStats represents the statistics for a URL.
type urlStats struct {
	AccessCount int64 `json:"access_count"`
}

// toURLStatsResponse converts an entity.URL to a urlStatsResponse.
func toURLStatsResponse(url *entity.URL) urlStatsResponse {
	return urlStatsResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		Stats: urlStats{
			AccessCount: url.URLStats.AccessCount,
		},
		CreatedAt: url.CreatedAt,
		UpdatedAt: url.UpdatedAt,
	}
}

// validationError represents an individual validation error.
type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// errorResponse represents a structured error response.
type errorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  []validationError `json:"errors,omitempty"`
}

// Predefined error responses for common scenarios.
var (
	emptyRequestBodyResponse = errorResponse{
		Status:  statusError,
		Message: "empty request body",
	}

	invalidRequestBodyResponse = errorResponse{
		Status:  statusError,
		Message: "invalid request body",
	}

	urlNotFoundResponse = errorResponse{
		Status:  statusError,
		Message: "url not found",
	}

	serverErrorResponse = errorResponse{
		Status:  statusError,
		Message: "server error occurred",
	}
)

// messageForTag returns a user-friendly message based on the validation tag.
func messageForTag(tag string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "url":
		return "invalid url"
	default:
		return "invalid value"
	}
}

// getValidationErrors processes validation errors and returns a list of validationError.
func getValidationErrors(err error) []validationError {
	var validationErrs []validationError

	errs, ok := err.(validator.ValidationErrors)
	if ok {
		for _, e := range errs {
			validationErrs = append(validationErrs, validationError{
				Field:   e.Field(),
				Message: messageForTag(e.Tag()),
			})
		}
	}

	return validationErrs
}

// validationErrorResponse constructs an errorResponse for validation errors.
func validationErrorResponse(err error) errorResponse {
	return errorResponse{
		Status:  statusError,
		Message: "validation error",
		Errors:  getValidationErrors(err),
	}
}
