package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type URLRepositoryTestSuite struct {
	suite.Suite
	pgCont  testcontainers.Container
	cfg     config.Postgres
	db      *sqlx.DB
	m       *migrate.Migrate
	urlRepo *postgres.URLRepository
}

func (suite *URLRepositoryTestSuite) SetupSuite() {
	ctx := context.Background()

	pgUser := "test"
	pgPassword := "test"
	pgDB := "url_shortener"

	var err error
	suite.pgCont, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/postgres:16-alpine",
			Env: map[string]string{
				"POSTGRES_USER":     pgUser,
				"POSTGRES_PASSWORD": pgPassword,
				"POSTGRES_DB":       pgDB,
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		err = suite.pgCont.Terminate(ctx)
		suite.Require().NoError(err)
	})

	pgHost, err := suite.pgCont.Host(ctx)
	suite.Require().NoError(err)

	pgPort, err := suite.pgCont.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	suite.cfg = config.Postgres{
		User:     pgUser,
		Password: pgPassword,
		Host:     pgHost,
		Port:     pgPort.Int(),
		DB:       pgDB,
		SSLMode:  "disable",
	}

	suite.db, err = sqlx.ConnectContext(ctx, "pgx", suite.cfg.DSN())
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		err := suite.db.Close()
		suite.Require().NoError(err)
	})

	migrationPath := "file://../../../../migrations"

	suite.m, err = migrate.New(migrationPath, suite.cfg.DSN())
	suite.Require().NoError(err)

	err = suite.m.Up()
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		err := suite.m.Down()
		suite.Require().NoError(err)
	})

	suite.urlRepo = postgres.NewURLRepository(suite.db)
}

func (suite *URLRepositoryTestSuite) SetupSubTest() {
	ctx := context.Background()

	_, err := suite.db.ExecContext(ctx, `TRUNCATE TABLE urls RESTART IDENTITY CASCADE`)
	suite.Require().NoError(err)
}

type urlRecord struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func insertURLRecord(t testing.TB, ctx context.Context, db *sqlx.DB, shortCode string, originalURL string) *urlRecord {
	t.Helper()

	rec := new(urlRecord)
	query := `INSERT INTO urls(short_code, original_url)
		VALUES ($1, $2)
		RETURNING *`

	if err := db.GetContext(ctx, rec, query, shortCode, originalURL); err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	return rec
}

func getURLRecord(t testing.TB, ctx context.Context, db *sqlx.DB, shortCode string) *urlRecord {
	t.Helper()

	rec := new(urlRecord)
	query := `SELECT * FROM urls
		WHERE short_code = $1`

	if err := db.GetContext(ctx, rec, query, shortCode); err != nil {
		t.Fatalf("Failed to get url record: %v", err)
	}

	return rec
}

func (suite *URLRepositoryTestSuite) Test_Create() {
	suite.Run("short code exists", func() {
		ctx := context.Background()
		_ = insertURLRecord(suite.T(), ctx, suite.db, "abc123", "https://example.com")

		url, err := suite.urlRepo.Create(ctx, "abc123", "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, database.ErrShortCodeExists)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		ctx := context.Background()
		url, err := suite.urlRepo.Create(ctx, "abc123", "https://example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)

		rec := getURLRecord(suite.T(), ctx, suite.db, "abc123")

		suite.Equal("abc123", rec.ShortCode)
		suite.Equal("https://example.com", rec.OriginalURL)
		suite.Zero(rec.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) Test_GetByShortCode() {
	suite.Run("url not found", func() {
		ctx := context.Background()
		url, err := suite.urlRepo.GetByShortCode(ctx, "abc123")

		suite.Error(err)
		suite.ErrorIs(err, database.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		ctx := context.Background()
		_ = insertURLRecord(suite.T(), ctx, suite.db, "abc123", "https://example.com")

		url, err := suite.urlRepo.GetByShortCode(ctx, "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Equal(int64(1), url.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) Test_Update() {
	suite.Run("url not found", func() {
		ctx := context.Background()
		url, err := suite.urlRepo.Update(ctx, "abc123", "https://new-example.com")

		suite.Error(err)
		suite.ErrorIs(err, database.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		ctx := context.Background()
		_ = insertURLRecord(suite.T(), ctx, suite.db, "abc123", "https://example.com")

		url, err := suite.urlRepo.Update(ctx, "abc123", "https://new-example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://new-example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)

		rec := getURLRecord(suite.T(), ctx, suite.db, "abc123")

		suite.Equal("abc123", rec.ShortCode)
		suite.Equal("https://new-example.com", url.OriginalURL)
		suite.Zero(rec.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) Test_Delete() {
	suite.Run("url not found", func() {
		ctx := context.Background()
		err := suite.urlRepo.Delete(ctx, "abc123")

		suite.Error(err)
		suite.ErrorIs(err, database.ErrURLNotFound)
	})

	suite.Run("success", func() {
		ctx := context.Background()
		_ = insertURLRecord(suite.T(), ctx, suite.db, "abc123", "https://example.com")

		err := suite.urlRepo.Delete(ctx, "abc123")

		suite.NoError(err)
	})
}

func (suite *URLRepositoryTestSuite) Test_GetStats() {
	suite.Run("url not found", func() {
		ctx := context.Background()
		url, err := suite.urlRepo.GetStats(ctx, "abc123")

		suite.Error(err)
		suite.ErrorIs(err, database.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		ctx := context.Background()
		_ = insertURLRecord(suite.T(), ctx, suite.db, "abc123", "https://example.com")

		url, err := suite.urlRepo.GetStats(ctx, "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)
	})
}

func TestURLRepository(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	suite.Run(t, new(URLRepositoryTestSuite))
}
