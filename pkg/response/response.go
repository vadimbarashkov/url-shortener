// Package response provides utilities for creating standardized
// API responses and handling validation errors in HTTP handlers.
package response

import "github.com/go-playground/validator/v10"

const (
	// StatusSuccess is a constant string representing a successful response status.
	StatusSuccess = "success"
	// StatusError is a constant string representing an error response status.
	StatusError = "error"
)

// Predefined error responses to be used across the application.
var (
	// EmptyRequestBodyResponse is returned when the request body is missing or empty.
	EmptyRequestBodyResponse = Response{
		Status:  StatusError,
		Message: "Request body is empty. Please provide necessary data.",
	}

	// BadRequestResponse is returned when the request body contains invalid data.
	BadRequestResponse = Response{
		Status:  StatusError,
		Message: "Invalid request body.",
	}

	// ResourceNotFoundResponse is returned when the requested resource is not found.
	ResourceNotFoundResponse = Response{
		Status:  StatusError,
		Message: "The requested resource was not found.",
	}

	// ServerErrorResponse is returned when an internal server error occurs.
	ServerErrorResponse = Response{
		Status:  StatusError,
		Message: "An internal server error occurred. Please try again later.",
	}
)

// Response represents a standardized API response.
type Response struct {
	// Status indicates the status of the response (e.g., "success" or "error")
	Status string `json:"status"`
	// Message contains a short description of the result or error.
	Message string `json:"message"`
	// Details holds any additional information, such as validation errors (optional).
	Details any `json:"details,omitempty"`
	// Data contains the result data for successful responses (optional).
	Data any `json:"data,omitempty"`
}

// SuccessResponse returns a successful API response with a message and optional data.
func SuccessResponse(msg string, data ...any) Response {
	resp := Response{
		Status:  StatusSuccess,
		Message: msg,
	}

	if len(data) > 0 && data[0] != nil {
		resp.Data = data[0]
	}

	return resp
}

// validationError represents a validation error for a specific field.
type validationError struct {
	// Field is the name of the field that failed validation.
	Field string `json:"field"`
	// Value is the value of the field that failed validation.
	Value any `json:"value"`
	// Issue provides a description of the validation issue.
	Issue string `json:"issue"`
}

// issueForTag maps validator tags to human-readable error messages.
func issueForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required."
	case "url":
		return "Invalid url."
	default:
		return "Invalid value."
	}
}

// getValidationErrors extracts validation errors from an error returned by the validator package.
func getValidationErrors(err error) []validationError {
	var validationErrs []validationError

	errs, ok := err.(validator.ValidationErrors)
	if ok {
		for _, e := range errs {
			validationErrs = append(validationErrs, validationError{
				Field: e.Field(),
				Value: e.Value(),
				Issue: issueForTag(e.Tag()),
			})
		}
	}

	return validationErrs
}

// ValidationErrorResponse creates a standardized error response for failed validation requests.
func ValidationErrorResponse(err error) Response {
	return Response{
		Status:  StatusError,
		Message: "Invalid request body. Please check your input.",
		Details: getValidationErrors(err),
	}
}
