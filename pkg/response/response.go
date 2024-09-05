package response

import "net/http"

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

var EmptyRequestBodyResponse = Response{
	Status:     StatusError,
	StatusCode: http.StatusBadRequest,
	Error:      "Empty Request Body",
	Message:    "Request body is empty. Please provide necessary data.",
}

var ResourseNotFoundResponse = Response{
	Status:     StatusError,
	StatusCode: http.StatusNotFound,
	Error:      "Resourse Not Found",
	Message:    "The requested resource was not found.",
}

var ServerErrorResponse = Response{
	Status:     StatusError,
	StatusCode: http.StatusInternalServerError,
	Error:      "Server Error",
	Message:    "An internal server error occurred. Please try again later.",
}

type Response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
	Message    string `json:"message"`
	Details    []any  `json:"details,omitempty"`
	Data       any    `json:"data,omitempty"`
}

func SuccessResponse(statusCode int, msg string, data ...any) Response {
	resp := Response{
		Status:     StatusSuccess,
		StatusCode: statusCode,
		Message:    msg,
	}

	if len(data) > 0 {
		resp.Data = data[0]
	}

	return resp
}
