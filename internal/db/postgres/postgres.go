package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(dsn string) (*sqlx.DB, error) {
	const op = "db.postgres.New"

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	return db, nil
}
