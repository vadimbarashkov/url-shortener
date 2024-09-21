package http

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

const statusError = "error"

type urlRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}

type urlResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toURLResponse(url *entity.URL) urlResponse {
	return urlResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}
}

type urlStatsResponse struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Stats       urlStats  `json:"stats"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type urlStats struct {
	AccessCount int64 `json:"access_count"`
}

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

type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  []validationError `json:"errors,omitempty"`
}

var emptyRequestBodyResponse = errorResponse{
	Status:  statusError,
	Message: "empty request body",
}

var invalidRequestBodyResponse = errorResponse{
	Status:  statusError,
	Message: "invalid request body",
}

var urlNotFoundResponse = errorResponse{
	Status:  statusError,
	Message: "url not found",
}

var serverErrorResponse = errorResponse{
	Status:  statusError,
	Message: "server error occurred",
}

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

func validationErrorResponse(err error) errorResponse {
	return errorResponse{
		Status:  statusError,
		Message: "validation error",
		Errors:  getValidationErrors(err),
	}
}
