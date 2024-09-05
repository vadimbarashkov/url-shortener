package response

const StatusSuccess = "success"

type Response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
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
