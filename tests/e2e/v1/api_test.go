package e2e

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/config"

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
	cfg *config.Config
	db  *sqlx.DB
	e   *httpexpect.Expect
}

func (suite *APITestSuite) SetupSuite() {
	root, err := findProjectRoot()
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

	baseURL := fmt.Sprintf("http://localhost:%d", suite.cfg.Server.Port)
	suite.e = httpexpect.Default(suite.T(), baseURL)
}

func (suite *APITestSuite) TearDownSuite() {
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
			Text().IsEqual("pong\n")
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

	suite.Run("invalid url value", func() {
		resp := suite.e.POST(path).
			WithJSON(map[string]string{
				"url": "invalid url",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("details").Array().Value(0).Object().
			HasValue("field", "url").
			HasValue("value", "invalid url").
			ContainsKey("issue")
	})

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
		resp.Value("data").Object().
			ContainsKey("id").
			ContainsKey("short_code").
			HasValue("url", "https://example.com").
			ContainsKey("created_at").
			ContainsKey("updated_at")
	})
}

func (suite *APITestSuite) TestResolveShortCode() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.GET(path, "abc123").
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		resp := suite.e.POST("/api/v1/shorten").
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		data := resp.Value("data").Object()
		shortCode := data.Value("short_code").String().Raw()

		resp = suite.e.GET(fmt.Sprintf(path, shortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
		resp.Value("data").Object().
			ContainsKey("id").
			HasValue("short_code", shortCode).
			HasValue("url", "https://example.com").
			ContainsKey("created_at").
			ContainsKey("updated_at")
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

	suite.Run("invalid url value", func() {
		resp := suite.e.PUT(fmt.Sprintf(path, "abc123")).
			WithJSON(map[string]string{
				"url": "invalid url",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
		resp.Value("details").Array().Value(0).Object().
			HasValue("field", "url").
			HasValue("value", "invalid url").
			ContainsKey("issue")
	})

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
		resp := suite.e.POST("/api/v1/shorten").
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		data := resp.Value("data").Object()
		shortCode := data.Value("short_code").String().Raw()

		resp = suite.e.PUT(fmt.Sprintf(path, shortCode)).
			WithJSON(map[string]string{
				"url": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
		resp.Value("data").Object().
			ContainsKey("id").
			HasValue("short_code", shortCode).
			HasValue("url", "https://new-example.com").
			ContainsKey("created_at").
			ContainsKey("updated_at")
	})
}

func (suite *APITestSuite) TestDeactivateURL() {
	const path = "/api/v1/shorten/%s"

	suite.Run("url not found", func() {
		resp := suite.e.DELETE(path, "abc123").
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		resp := suite.e.POST("/api/v1/shorten").
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		data := resp.Value("data").Object()
		shortCode := data.Value("short_code").String().Raw()

		resp = suite.e.DELETE(fmt.Sprintf(path, shortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
	})
}

func (suite *APITestSuite) TestGetURLStats() {
	const path = "/api/v1/shorten/%s/stats"

	suite.Run("url not found", func() {
		resp := suite.e.GET(path, "abc123").
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", "error")
		resp.ContainsKey("message")
	})

	suite.Run("success", func() {
		resp := suite.e.POST("/api/v1/shorten").
			WithJSON(map[string]string{
				"url": "https://example.com",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		data := resp.Value("data").Object()
		shortCode := data.Value("short_code").String().Raw()

		resp = suite.e.GET(fmt.Sprintf(path, shortCode)).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("status", "success")
		resp.ContainsKey("message")
		resp.Value("data").Object().
			ContainsKey("id").
			HasValue("short_code", shortCode).
			HasValue("url", "https://example.com").
			HasValue("access_count", int64(0)).
			ContainsKey("created_at").
			ContainsKey("updated_at")
	})
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
