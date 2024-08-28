package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const uniqueViolationErrCode = "23505"

func isUniqueViolationError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.SQLState() == uniqueViolationErrCode
}

func New(dsn string) (*sqlx.DB, error) {
	const op = "database.postgres.New"

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	return db, nil
}
