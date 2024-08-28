package config

import "time"

type Config struct {
	Env string
	Server
	Postgres
}

type Server struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	CertFile     string
	KeyFile      string
}

type Postgres struct {
	User     string
	Password string
	Host     string
	Port     int
	DB       string
	SSLMode  string
}
