package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/httplog/v2"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vadimbarashkov/url-shortener/internal/adapter/repository/postgres"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/internal/usecase"
	"github.com/vadimbarashkov/url-shortener/tests"

	delivery "github.com/vadimbarashkov/url-shortener/internal/adapter/delivery/http"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type APITestSuite struct {
	suite.Suite
	pgCont     testcontainers.Container
	cfg        config.Postgres
	db         *sqlx.DB
	urlRepo    *postgres.URLRepository
	urlUseCase *usecase.URLUseCase
	logger     *httplog.Logger
	server     *httptest.Server
	e          *httpexpect.Expect
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

	root, err := tests.FindProjectRoot()
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
	suite.urlUseCase = usecase.NewURLUseCase(suite.urlRepo)

	suite.logger = httplog.NewLogger("", httplog.Options{Writer: io.Discard})
	router := delivery.NewRouter(suite.logger, suite.urlUseCase)
	suite.server = httptest.NewServer(router)
	suite.e = httpexpect.Default(suite.T(), suite.server.URL)
}

func (suite *APITestSuite) TearDownSubTest() {
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
			Text().IsEqual("pong")
	})
}

func (suite *APITestSuite) TestShortenURL() {
	const path = "/api/v1/shorten"

	suite.Run("success", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "https://example.com"}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		shortCode := resp.Value("short_code").String().Raw()

		url, err := suite.urlRepo.RetrieveByShortCode(context.Background(), shortCode)
		if err != nil {
			suite.T().Fatalf("Failed to retrieve url record: %v", err)
		}

		resp.HasValue("id", url.ID)
		resp.HasValue("short_code", url.ShortCode)
		resp.HasValue("original_url", url.OriginalURL)
		resp.NotContainsKey("stats")
		resp.HasValue("created_at", url.CreatedAt)
		resp.ContainsKey("updated_at")
	})
}

func (suite *APITestSuite) TestResolveShortCode() {
	path := "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		url, err := suite.urlRepo.Save(context.Background(), "abc123", "https://example.com")
		if err != nil {
			suite.T().Fatalf("Failed to save url record: %v", err)
		}

		resp := suite.e.GET(fmt.Sprintf(path, url.ShortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("id", url.ID)
		resp.HasValue("short_code", url.ShortCode)
		resp.HasValue("original_url", url.OriginalURL)
		resp.NotContainsKey("stats")
		resp.HasValue("created_at", url.CreatedAt)
		resp.ContainsKey("updated_at")

		url, err = suite.urlRepo.RetrieveByShortCode(context.Background(), url.ShortCode)
		if err != nil {
			suite.T().Fatalf("Failed to retrieve url record: %v", err)
		}

		suite.Equal(int64(1), url.AccessCount)
	})
}

func (suite *APITestSuite) TestModifyURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "https://new-example.com"}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		url, err := suite.urlRepo.Save(context.Background(), "abc123", "https://example.com")
		if err != nil {
			suite.T().Fatalf("Failed to save url record: %v", err)
		}

		resp := suite.e.PUT(fmt.Sprintf(path, url.ShortCode)).
			WithJSON(map[string]string{"original_url": "https://new-example.com"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("id", url.ID)
		resp.HasValue("short_code", url.ShortCode)
		resp.HasValue("original_url", "https://new-example.com")
		resp.NotContainsKey("stats")
		resp.HasValue("created_at", url.CreatedAt)
		resp.ContainsKey("updated_at")
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
		url, err := suite.urlRepo.Save(context.Background(), "abc123", "https://example.com")
		if err != nil {
			suite.T().Fatalf("Failed to save url record: %v", err)
		}

		suite.e.DELETE(fmt.Sprintf(path, url.ShortCode)).
			Expect().
			Status(http.StatusNoContent)
	})
}

func (suite *APITestSuite) TestGetURLStats() {
	path := "/api/v1/shorten/%s/stats"

	suite.Run("url not found", func() {
		resp := suite.e.GET(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		url, err := suite.urlRepo.Save(context.Background(), "abc123", "https://example.com")
		if err != nil {
			suite.T().Fatalf("Failed to save url record: %v", err)
		}

		resp := suite.e.GET(fmt.Sprintf(path, url.ShortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("id", url.ID)
		resp.HasValue("short_code", url.ShortCode)
		resp.HasValue("original_url", url.OriginalURL)
		resp.Value("stats").Object().
			HasValue("access_count", int64(0))
		resp.HasValue("created_at", url.CreatedAt)
		resp.ContainsKey("updated_at")
	})
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
