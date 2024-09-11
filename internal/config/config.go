package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config represents the project configuration.
type Config struct {
	Env             string `env:"ENV" envDefault:"dev"`
	ShortCodeLength int    `env:"SHORT_CODE_LENGTH" envDefault:"7"`
	Server          `envPrefix:"SERVER_"`
	Postgres        `envPrefix:"POSTGRES_"`
}

// Load loads the project configuration from the specified .env file path.
// It returns a pointer to Config struct or an error.
func Load(path string) (*Config, error) {
	const op = "config.Load"

	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("%s: failed to load .env file: %w", op, err)
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse envs: %w", op, err)
	}

	return &cfg, nil
}

// Server represents the server configuration.
type Server struct {
	Port         int           `env:"PORT" envDefault:"8443"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"1s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"2s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"1m"`
	CertFile     string        `env:"CERT_FILE"`
	KeyFile      string        `env:"KEY_FILE"`
}

// Addr returns the server address in the format ":<port>".
func (s *Server) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

// Postgres represents the PostgreSQL configuration.
type Postgres struct {
	User     string `env:"USER" envDefault:"postgres"`
	Password string `env:"PASSWORD"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	DB       string `env:"DB"`
	SSLMode  string `env:"SSLMODE" envDefault:"disable"`
}

// DSN returns the Data Source Name (DSN) for connecting to the PostgreSQL.
func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}
