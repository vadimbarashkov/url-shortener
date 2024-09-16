package api

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/httplog/v2"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/database/postgres"
	"github.com/vadimbarashkov/url-shortener/internal/service"
	"github.com/vadimbarashkov/url-shortener/pkg/response"

	api "github.com/vadimbarashkov/url-shortener/internal/api/http/v1"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

type APITestSuite struct {
	suite.Suite
	pgCont  testcontainers.Container
	cfg     config.Postgres
	db      *sqlx.DB
	urlRepo *postgres.URLRepository
	urlSvc  *service.URLService
	logger  *httplog.Logger
	server  *httptest.Server
	e       *httpexpect.Expect
}

func (suite *APITestSuite) SetupSuite() {
	ctx := context.Background()

	pgUser := "test"
	pgPassword := "test"
	pgDB := "url_shortener"

	var err error
	suite.pgCont, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:16-alpine",
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
	if err != nil {
		suite.T().Fatalf("Failed to start postgres container: %v", err)
	}
	suite.T().Cleanup(func() {
		if err := suite.pgCont.Terminate(ctx); err != nil {
			suite.T().Fatalf("Failed to terminate postgres container: %v", err)
		}
	})

	pgHost, err := suite.pgCont.Host(ctx)
	if err != nil {
		suite.T().Fatalf("Failed to get postgres container host: %v", err)
	}

	pgPort, err := suite.pgCont.MappedPort(ctx, "5432")
	if err != nil {
		suite.T().Fatalf("Failed to get postgres container port: %v", err)
	}

	suite.cfg = config.Postgres{
		User:     pgUser,
		Password: pgPassword,
		Host:     pgHost,
		Port:     pgPort.Int(),
		DB:       pgDB,
		SSLMode:  "disable",
	}

	suite.db, err = sqlx.Connect("pgx", suite.cfg.DSN())
	if err != nil {
		suite.T().Fatalf("Failed to connect to database: %v", err)
	}
	suite.T().Cleanup(func() {
		if err := suite.db.Close(); err != nil {
			suite.T().Fatalf("Failed to close database: %v", err)
		}
	})

	root, err := findProjectRoot()
	if err != nil {
		suite.T().Fatalf("Failed to get project root: %v", err)
	}

	migrationsPath := filepath.Join("file://"+root, "/migrations")

	m, err := migrate.New(migrationsPath, suite.cfg.DSN())
	if err != nil {
		suite.T().Fatalf("Failed to initialize migrations: %v", err)
	}

	if err := m.Up(); err != nil {
		suite.T().Fatalf("Failed to run migrations: %v", err)
	}
	suite.T().Cleanup(func() {
		if err := m.Down(); err != nil {
			suite.T().Fatalf("Failed to rollback migrations: %v", err)
		}
	})

	suite.urlRepo = postgres.NewURLRepository(suite.db)
	suite.urlSvc = service.NewURLService(suite.urlRepo, 7)

	suite.logger = httplog.NewLogger("", httplog.Options{Writer: io.Discard})
	router := api.NewRouter(suite.logger, suite.urlSvc)
	suite.server = httptest.NewServer(router)
	suite.e = httpexpect.Default(suite.T(), suite.server.URL)
}

func (suite *APITestSuite) SetupSubTest() {
	ctx := context.Background()

	_, err := suite.db.ExecContext(ctx, `TRUNCATE TABLE urls RESTART IDENTITY CASCADE`)
	if err != nil {
		suite.T().Fatalf("Failed to clean urls table: %v", err)
	}
}

func (suite *APITestSuite) TestPing() {
	const path = "/api/v1/ping"

	suite.Run("success", func() {
		suite.e.GET(path).
			Expect().
			Status(http.StatusOK).
			Text().IsEqual("pong\n")
	})
}

type urlRecord struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func insertURLRecord(t testing.TB, db *sqlx.DB, shortCode, originalURL string) *urlRecord {
	t.Helper()

	rec := new(urlRecord)
	query := `INSERT INTO urls(short_code, original_url)
		VALUES ($1, $2)
		RETURNING *`

	if err := db.Get(rec, query, shortCode, originalURL); err != nil {
		t.Fatalf("Failed to insert url record: %v", err)
	}

	return rec
}

func getURLRecord(t testing.TB, db *sqlx.DB, shortCode string) *urlRecord {
	t.Helper()

	rec := new(urlRecord)
	query := `SELECT * FROM urls
		WHERE short_code = $1`

	if err := db.Get(rec, query, shortCode); err != nil {
		t.Fatalf("Failed to get url record: %v", err)
	}

	return rec
}

func (suite *APITestSuite) TestShortenURL() {
	const path = "/api/v1/shorten"

	suite.Run("success", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")

		data := resp.Value("data").Object()
		rec := getURLRecord(suite.T(), suite.db, data.Value("short_code").String().Raw())

		data.HasValue("id", rec.ID).
			HasValue("short_code", rec.ShortCode).
			HasValue("url", rec.OriginalURL).
			HasValue("created_at", rec.CreatedAt).
			ContainsKey("updated_at")
	})
}

func (suite *APITestSuite) TestResolveShortCode() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		rec := insertURLRecord(suite.T(), suite.db, "abc123", "https://example.com")

		resp := suite.e.GET(fmt.Sprintf(path, rec.ShortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
		resp.Value("data").Object().
			HasValue("id", rec.ID).
			HasValue("short_code", rec.ShortCode).
			HasValue("url", rec.OriginalURL).
			HasValue("created_at", rec.CreatedAt).
			ContainsKey("updated_at")

		rec = getURLRecord(suite.T(), suite.db, rec.ShortCode)

		suite.Equal(rec.AccessCount, int64(1))
	})
}

func (suite *APITestSuite) TestModifyURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		rec := insertURLRecord(suite.T(), suite.db, "abc123", "https://example.com")

		resp := suite.e.PUT(fmt.Sprintf(path, rec.ShortCode)).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
		resp.Value("data").Object().
			HasValue("id", rec.ID).
			HasValue("short_code", rec.ShortCode).
			HasValue("url", "https://new-example.com").
			HasValue("created_at", rec.CreatedAt).
			ContainsKey("updated_at")

		rec = getURLRecord(suite.T(), suite.db, "abc123")

		suite.Equal("https://new-example.com", rec.OriginalURL)
	})
}

func (suite *APITestSuite) TestDeactivateURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		insertURLRecord(suite.T(), suite.db, "abc123", "https://example.com")

		resp := suite.e.DELETE(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", response.StatusSuccess)
		resp.ContainsKey("message")

		rec := new(urlRecord)
		query := `SELECT * FROM urls
			WHERE short_code = $1`

		err := suite.db.Get(rec, query, "abc123")
		suite.Error(err)
		suite.ErrorIs(err, sql.ErrNoRows)
	})
}

func (suite *APITestSuite) TestGetURLStats() {
	const path = "/api/v1/shorten/%s/stats"

	suite.Run("url not found", func() {
		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		rec := new(urlRecord)
		query := `INSERT INTO urls(short_code, original_url, access_count)
			VALUES ($1, $2, $3)
			RETURNING *`

		if err := suite.db.Get(rec, query, "abc123", "https://example.com", 1); err != nil {
			suite.T().Fatalf("Failed to insert url record: %v", err)
		}

		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", response.StatusSuccess)
		resp.ContainsKey("message")
		resp.Value("data").Object().
			HasValue("id", rec.ID).
			HasValue("short_code", rec.ShortCode).
			HasValue("url", rec.OriginalURL).
			HasValue("access_count", rec.AccessCount).
			HasValue("created_at", rec.CreatedAt).
			ContainsKey("updated_at")
	})
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
