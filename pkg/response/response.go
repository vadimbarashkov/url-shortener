package response

import "github.com/go-playground/validator/v10"

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

var EmptyRequestBodyResponse = Response{
	Status:  StatusError,
	Message: "Request body is empty. Please provide necessary data.",
}

var ResourceNotFoundResponse = Response{
	Status:  StatusError,
	Message: "The requested resource was not found.",
}

var ServerErrorResponse = Response{
	Status:  StatusError,
	Message: "An internal server error occurred. Please try again later.",
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Data    any    `json:"data,omitempty"`
}

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

func ValidationErrorResponse(err error) Response {
	return Response{
		Status:  StatusError,
		Message: "Invalid request body. Please check your input.",
		Details: getValidationErrors(err),
	}
}

type validationError struct {
	Field string `json:"field"`
	Value any    `json:"value"`
	Issue string `json:"issue"`
}

func issueForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required."
	case "url":
		return "Invalid url."
	default:
		return ""
	}
}

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
