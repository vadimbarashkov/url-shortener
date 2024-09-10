package response

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		data []any
		want Response
	}{
		{
			name: "without data",
			msg:  "Operation successful.",
			want: Response{
				Status:  StatusSuccess,
				Message: "Operation successful.",
			},
		},
		{
			name: "with data",
			msg:  "Operation successful.",
			data: []any{map[string]any{"id": 1}},
			want: Response{
				Status:  StatusSuccess,
				Message: "Operation successful.",
				Data:    map[string]any{"id": 1},
			},
		},
		{
			name: "with multiple data",
			msg:  "Operation successful.",
			data: []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
			want: Response{
				Status:  StatusSuccess,
				Message: "Operation successful.",
				Data:    map[string]any{"id": 1},
			},
		},
		{
			name: "with nil data",
			msg:  "Operation successful.",
			data: nil,
			want: Response{
				Status:  StatusSuccess,
				Message: "Operation successful.",
			},
		},
		{
			name: "with data containing nil",
			msg:  "Operation successful.",
			data: []any{nil},
			want: Response{
				Status:  StatusSuccess,
				Message: "Operation successful.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuccessResponse(tt.msg, tt.data...)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetValidationErrors(t *testing.T) {
	type req struct {
		Name string `json:"name" validate:"required"`
		URL  string `json:"url" validate:"required,url"`
	}

	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	tests := []struct {
		name string
		req  req
		want []validationError
	}{
		{
			name: "not validation error",
			req: req{
				Name: "name",
				URL:  "https://example.com",
			},
		},
		{
			name: "one error",
			req: req{
				Name: "",
				URL:  "https://example.com",
			},
			want: []validationError{
				{
					Field: "name",
					Value: "",
					Issue: "This field is required.",
				},
			},
		},
		{
			name: "two errors",
			req: req{
				Name: "",
				URL:  "not url",
			},
			want: []validationError{
				{
					Field: "name",
					Value: "",
					Issue: "This field is required.",
				},
				{
					Field: "url",
					Value: "not url",
					Issue: "Invalid url.",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.req)
			got := getValidationErrors(err)

			assert.Equal(t, tt.want, got)
		})
	}
}
