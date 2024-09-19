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
)

type Config struct {
	Env             string `yaml:"env"`
	ShortCodeLength int    `yaml:"short_code_length"`
	HTTPServer      `yaml:"http_server"`
	Postgres        `yaml:"postgres"`
}

type HTTPServer struct {
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	CertFile       string        `yaml:"cert_file"`
	KeyFile        string        `yaml:"key_file"`
}

var defaultHTTPServer = HTTPServer{
	Port:           8080,
	ReadTimeout:    5 * time.Second,
	WriteTimeout:   10 * time.Second,
	IdleTimeout:    time.Minute,
	MaxHeaderBytes: 1 << 20,
}

func (s *HTTPServer) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

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

var defaultPostgres = Postgres{
	Host:            "localhost",
	Port:            5432,
	SSLMode:         "disable",
	ConnMaxIdleTime: 5 * time.Minute,
	ConnMaxLifetime: 30 * time.Minute,
	MaxIdleConns:    5,
	MaxOpenConns:    25,
}

func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}

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

func setDefaults(cfg *Config) {
	cfg.Env = EnvDev
	cfg.ShortCodeLength = 7
	cfg.HTTPServer = defaultHTTPServer
	cfg.Postgres = defaultPostgres
}
