package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupPostgres(t testing.TB) config.Postgres {
	t.Helper()

	ctx := context.Background()

	pgUser := "test"
	pgPassword := "test"
	pgDB := "url_shortener"

	pgCont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:16-alpine",
			Env: map[string]string{
				"POSTGRES_USER":     pgUser,
				"POSTGRES_PASSWORD": pgPassword,
				"POSTGRES_DB":       pgDB,
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForExposedPort(),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgCont.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate postgres container: %v", err)
		}
	})

	pgHost, err := pgCont.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}
	pgPort, err := pgCont.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	return config.Postgres{
		User:     pgUser,
		Password: pgPassword,
		Host:     pgHost,
		Port:     pgPort.Int(),
		DB:       pgDB,
		SSLMode:  "disable",
	}
}

func runMigrations(t testing.TB, cfg config.Postgres) {
	t.Helper()

	migrationPath := "file://../../../../migrations"

	m, err := migrate.New(migrationPath, cfg.DSN())
	if err != nil {
		t.Fatalf("Failed to initialize migrations: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("Failed to rollback migrations: %v", err)
		}
	})
}

func setupURLRepository(t testing.TB) (*postgres.URLRepository, *sqlx.DB) {
	t.Helper()

	cfg := setupPostgres(t)
	runMigrations(t, cfg)

	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
	})

	return postgres.NewURLRepository(db), db
}

type urlRecord struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func TestURLRepository_Create(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("duplicate short_code", func(t *testing.T) {
		repo, _ := setupURLRepository(t)

		url, err := repo.Create(context.Background(), "abc123", "https://example.com")

		assert.NoError(t, err)
		assert.NotNil(t, url)

		url, err = repo.Create(context.Background(), "abc123", "https://example2.com")

		assert.Error(t, err)
		assert.Error(t, err, database.ErrShortCodeExists)
		assert.Nil(t, url)
	})

	t.Run("success", func(t *testing.T) {
		repo, db := setupURLRepository(t)

		url, err := repo.Create(context.Background(), "abc123", "https://example.com")

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, "abc123", url.ShortCode)
		assert.Equal(t, "https://example.com", url.OriginalURL)
		assert.Zero(t, url.AccessCount)

		var rec urlRecord
		query := `SELECT * FROM urls
			WHERE short_code = $1`

		if err := db.GetContext(context.Background(), &rec, query, "abc123"); err != nil {
			t.Fatalf("Failed to get url record: %v", err)
		}

		assert.Equal(t, "abc123", rec.ShortCode)
		assert.Equal(t, "https://example.com", rec.OriginalURL)
		assert.Zero(t, rec.AccessCount)
	})
}
