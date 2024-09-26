package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/adapter/repository/postgres"
	"github.com/vadimbarashkov/url-shortener/internal/config"
	"github.com/vadimbarashkov/url-shortener/tests"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type APITestSuite struct {
	suite.Suite
	cfg     *config.Config
	db      *sqlx.DB
	urlRepo *postgres.URLRepository
	e       *httpexpect.Expect
}

func (suite *APITestSuite) SetupSuite() {
	root, err := tests.FindProjectRoot()
	if err != nil {
		suite.T().Fatalf("Failed to get project root: %v", err)
	}

	configPath := filepath.Join(root, os.Getenv("CONFIG_PATH"))

	suite.cfg, err = config.Load(configPath)
	if err != nil {
		suite.T().Fatalf("Failed to load config: %v", err)
	}

	suite.db, err = sqlx.Connect("pgx", suite.cfg.Postgres.DSN())
	if err != nil {
		suite.T().Fatalf("Failed to connect to database: %v", err)
	}
	suite.T().Cleanup(func() {
		suite.db.Close()
	})

	suite.urlRepo = postgres.NewURLRepository(suite.db)

	baseURL := fmt.Sprintf("http://localhost:%d", suite.cfg.HTTPServer.Port)
	suite.e = httpexpect.Default(suite.T(), baseURL)
}

func (suite *APITestSuite) TearDownSubTest() {
	_, err := suite.db.Exec(`TRUNCATE TABLE urls RESTART IDENTITY CASCADE`)
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

	suite.Run("empty request body", func() {
		resp := suite.e.POST(path).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("invalid request body", func() {
		resp := suite.e.POST(path).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("validation error", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "invalid url"}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("errors").Array().Value(0).Object().
			HasValue("field", "original_url").
			ContainsKey("message")
	})

	suite.Run("success", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{"original_url": "https://example.com"}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.ContainsKey("id")
		resp.ContainsKey("short_code")
		resp.HasValue("original_url", "https://example.com")
		resp.NotContainsKey("stats")
		resp.ContainsKey("created_at")
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
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func (suite *APITestSuite) TestModifyURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("empty request body", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("invalid request body", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("validation error", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{"original_url": "invalid url"}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("errors").Array().Value(0).Object().
			HasValue("field", "original_url").
			ContainsKey("message")
	})

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
		resp.ContainsKey("created_at")
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
		resp.ContainsKey("created_at")
		resp.ContainsKey("updated_at")
	})
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
