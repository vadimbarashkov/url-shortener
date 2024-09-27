// Package config provides functionality for loading and managing application configuration
// from a YAML file, with support for setting default values.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	EnvDev   = "dev"
	EnvStage = "stage"
	EnvProd  = "prod"

	defaultShortCodeLength = 7
)

// Config represents the application's configuration.
type Config struct {
	Env             string `yaml:"env"`
	ShortCodeLength int    `yaml:"short_code_length"`
	HTTPServer      `yaml:"http_server"`
	Postgres        `yaml:"postgres"`
}

// HTTPServer contains the configuration for the HTTP server.
type HTTPServer struct {
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	CertFile       string        `yaml:"cert_file"`
	KeyFile        string        `yaml:"key_file"`
}

// defaultHTTPServer holds the default settings for the HTTP server.
var defaultHTTPServer = HTTPServer{
	Port:           8080,
	ReadTimeout:    5 * time.Second,
	WriteTimeout:   10 * time.Second,
	IdleTimeout:    time.Minute,
	MaxHeaderBytes: 1 << 20,
}

// Addr returns the address the HTTP server will bind to, formatted as <:port>.
func (s *HTTPServer) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

// Postgres contains PostgreSQL database connection settings.
type Postgres struct {
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	DB              string        `yaml:"db"`
	SSLMode         string        `yaml:"sslmode"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
}

// defaultPostgres holds the default settings for PostgreSQL connection.
var defaultPostgres = Postgres{
	Host:            "localhost",
	Port:            5432,
	SSLMode:         "disable",
	ConnMaxIdleTime: 5 * time.Minute,
	ConnMaxLifetime: 30 * time.Minute,
	MaxIdleConns:    5,
	MaxOpenConns:    25,
}

// DSN returns a PostgreSQL Data Source Name (DSN) string for connecting to the PostgreSQL database.
func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}

// Load reads a configuration YAML file from the specified path and loads it into a Config struct.
// If any fields are missing from the file, default values are assigned using the setDefaults function.
// It returns a pointer to the Config struct and an error if the loading process fails.
func Load(path string) (*Config, error) {
	const op = "config.Load"

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open config file: %w", op, err)
	}
	defer f.Close()

	var cfg Config
	setDefaults(&cfg)

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to decode config file: %w", op, err)
	}

	return &cfg, nil
}

// setDefaults applies default values to the Config struct.
func setDefaults(cfg *Config) {
	cfg.Env = EnvDev
	cfg.ShortCodeLength = defaultShortCodeLength
	cfg.HTTPServer = defaultHTTPServer
	cfg.Postgres = defaultPostgres
}
