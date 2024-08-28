package config

import (
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func createEnvFile(t testing.TB, data []byte) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})

	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}

	return f
}

func TestLoad(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		cfg, err := Load("invalid/path/to/.env")

		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, cfg)
	})

	t.Run("invalid .env file", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		data := `ENV=test
SERVER_PORT=not_number
SERVER_READ_TIMEOUT=1s
SERVER_WRITE_TIMEOUT=2s
SERVER_IDLE_TIMEOUT=1m
SERVER_CERT_FILE=./example.pem
SERVER_KEY_FILE=./example-key.pem
POSTGRES_USER=test_user
POSTGRES_PASSWORD=test_pass
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=test_db
POSTGRES_SSLMODE=disable`

		f := createEnvFile(t, []byte(data))
		cfg, err := Load(f.Name())

		assert.Error(t, err)
		assert.ErrorIs(t, err, env.ParseError{})
		assert.Nil(t, cfg)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		data := `ENV=test
SERVER_PORT=8443
SERVER_READ_TIMEOUT=1s
SERVER_WRITE_TIMEOUT=2s
SERVER_IDLE_TIMEOUT=1m
SERVER_CERT_FILE=./example.pem
SERVER_KEY_FILE=./example-key.pem
POSTGRES_USER=test_user
POSTGRES_PASSWORD=test_pass
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=test_db
POSTGRES_SSLMODE=disable`

		wantCfg := Config{
			Env: "test",
			Server: Server{
				Port:         8443,
				ReadTimeout:  time.Second,
				WriteTimeout: 2 * time.Second,
				IdleTimeout:  time.Minute,
				CertFile:     "./example.pem",
				KeyFile:      "./example-key.pem",
			},
			Postgres: Postgres{
				User:     "test_user",
				Password: "test_pass",
				Host:     "localhost",
				Port:     5432,
				DB:       "test_db",
				SSLMode:  "disable",
			},
		}

		f := createEnvFile(t, []byte(data))
		cfg, err := Load(f.Name())

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, wantCfg, *cfg)
	})
}
