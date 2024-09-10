package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/vadimbarashkov/url-shortener/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const uniqueViolationErrCode = "23505"

// isUniqueViolationError checks if the given error is a PostgreSQL unique
// violation error, which occurs when a duplicate value is inserted into a
// column with a unique constraint. It returns true if the error is a
// unique violation, otherwise false.
func isUniqueViolationError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.SQLState() == uniqueViolationErrCode
}

// New creates a new connection to the PostgreSQL database using the provided
// Data Source Name (DSN). It returns a pointer to sqlx.DB object representing
// the connection, or an error if the connection could not be established.
func New(cfg config.Postgres) (*sqlx.DB, error) {
	const op = "database.postgres.New"

	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	return db, nil
}
