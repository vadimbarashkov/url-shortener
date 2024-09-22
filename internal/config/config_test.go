package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("non-existent config file", func(t *testing.T) {
		cfg, err := Load("invalid/path/to/config.yml")

		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, cfg)
	})

	t.Run("invalid config file", func(t *testing.T) {
		data := `http_server:
  port: not number
  cert_file: ./crts/example.pem
  key_file: ./crts/example-key.pem
postgres:
  user: test
  password: test
  db: test`

		f := createTempFile(t, []byte(data))
		cfg, err := Load(f.Name())

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("success", func(t *testing.T) {
		data := `http_server:
  cert_file: ./crts/example.pem
  key_file: ./crts/example-key.pem
postgres:
  user: test
  password: test
  db: test`

		f := createTempFile(t, []byte(data))
		cfg, err := Load(f.Name())

		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		var wantCfg Config
		setDefaults(&wantCfg)

		wantCfg.HTTPServer.CertFile = "./crts/example.pem"
		wantCfg.HTTPServer.KeyFile = "./crts/example-key.pem"
		wantCfg.Postgres.User = "test"
		wantCfg.Postgres.Password = "test"
		wantCfg.Postgres.DB = "test"

		assert.Equal(t, wantCfg, *cfg)
	})
}

func createTempFile(t testing.TB, data []byte) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", "config.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})

	if _, err := f.Write(data); err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	return f
}

func TestHTTPServer_Addr(t *testing.T) {
	s := HTTPServer{Port: 8080}

	assert.Equal(t, ":8080", s.Addr())
}

func TestPostgres_DSN(t *testing.T) {
	p := Postgres{
		User:     "test",
		Password: "test",
		Host:     "localhost",
		Port:     5432,
		DB:       "test",
		SSLMode:  "disable",
	}

	assert.Equal(t, "postgres://test:test@localhost:5432/test?sslmode=disable", p.DSN())
}
