package config_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/url-shortener/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		path := "invalid/path/to/config.yml"

		cfg, err := config.Load(path)

		assert.Nil(t, cfg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("invalid config file", func(t *testing.T) {
		f, err := os.CreateTemp("", "config.yml")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())

		data := ``

		if _, err := f.WriteString(data); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(f.Name())

		assert.Nil(t, cfg)
		assert.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("valid config file", func(t *testing.T) {
		f, err := os.CreateTemp("", "config.yml")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())

		data := `env: test
server:
  port: 8080
  read_timeout: 1s
  write_timeout: 2s
  idle_timeout: 60s
  max_header_bytes: 1048576
redis:
  password:
  host: localhost
  port: 6379
  database: 0`

		if _, err := f.WriteString(data); err != nil {
			t.Fatal(err)
		}

		expectedCfg := config.Config{
			Env: "test",
			Server: config.Server{
				Port:           8080,
				ReadTimeout:    time.Second,
				WriteTimeout:   2 * time.Second,
				IdleTimeout:    60 * time.Second,
				MaxHeaderBytes: 1 << 20,
			},
			Redis: config.Redis{
				Password: "",
				Host:     "localhost",
				Port:     6379,
				Database: 0,
			},
		}

		cfg, err := config.Load(f.Name())

		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.Equal(t, expectedCfg, *cfg)
	})
}

func TestServer_Addr(t *testing.T) {
	s := config.Server{
		Port:           8080,
		ReadTimeout:    time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	expectedAddr := ":8080"

	assert.Equal(t, expectedAddr, s.Addr())
}

func TestRedis_Addr(t *testing.T) {
	r := config.Redis{
		Password: "",
		Host:     "localhost",
		Port:     6379,
		Database: 0,
	}

	expectedAddr := "localhost:6379"

	assert.Equal(t, expectedAddr, r.Addr())
}
