package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique violation error",
			err:  &pgconn.PgError{Code: uniqueViolationErrCode},
			want: true,
		},
		{
			name: "not unique violation error",
			err:  &pgconn.PgError{Code: "unknown error code"},
			want: false,
		},
		{
			name: "not PgError",
			err:  errors.New("unknown error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolationError(tt.err)

			assert.Equal(t, tt.want, got)
		})
	}
}
