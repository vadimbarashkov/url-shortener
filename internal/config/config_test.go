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

func TestMustLoad(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		defer func() {
			err, _ := recover().(error)

			assert.Error(t, err)
			assert.ErrorIs(t, err, os.ErrNotExist)
		}()

		MustLoad("invalid/paht/to/.env")
	})

	t.Run("invalid .env file", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		defer func() {
			err, _ := recover().(error)

			assert.Error(t, err)
			assert.ErrorIs(t, err, env.ParseError{})
		}()

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
		MustLoad(f.Name())
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		defer func() {
			err, _ := recover().(error)

			assert.NoError(t, err)
		}()

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
			Env:             "test",
			ShortCodeLength: 7,
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
		cfg := MustLoad(f.Name())

		assert.NotNil(t, cfg)
		assert.Equal(t, wantCfg, *cfg)
	})
}

func TestServer_Addr(t *testing.T) {
	s := Server{Port: 8443}

	wantAddr := ":8443"

	assert.Equal(t, wantAddr, s.Addr())
}

func TestPostgres_DSN(t *testing.T) {
	p := Postgres{
		User:     "test_user",
		Password: "test_pass",
		Host:     "localhost",
		Port:     5432,
		DB:       "test_db",
		SSLMode:  "disable",
	}

	wantDSN := "postgres://test_user:test_pass@localhost:5432/test_db?sslmode=disable"

	assert.Equal(t, wantDSN, p.DSN())
}
