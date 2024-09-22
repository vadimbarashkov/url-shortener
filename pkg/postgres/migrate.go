package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies the database migrations from the specified path using the provided Data Source Name (DSN).
func RunMigrations(path string, dsn string) error {
	const op = "postgres.RunMigrations"

	m, err := migrate.New(path, dsn)
	if err != nil {
		return fmt.Errorf("%s: failed to initialize migrations: %w", op, err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}

	return nil
}
