package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env    string `yaml:"env"`
	Server `yaml:"server"`
	Redis  `yaml:"redis"`
}

func Load(path string) (*Config, error) {
	const op = "config.Load"

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open config file: %w", op, err)
	}
	defer f.Close()

	var cfg Config

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to decode config file: %w", op, err)
	}

	return &cfg, nil
}

type Server struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

func (s Server) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

type Redis struct {
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database int    `yaml:"database"`
}

func (r Redis) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
