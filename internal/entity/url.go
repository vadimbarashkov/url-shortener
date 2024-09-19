package entity

import (
	"errors"
	"time"
)

var (
	ErrShortCodeExists = errors.New("short code exists")
	ErrURLNotFound     = errors.New("url not found")
)

type URL struct {
	ID          int64
	ShortCode   string
	OriginalURL string
	URLStats
	CreatedAt time.Time
	UpdatedAt time.Time
}

type URLStats struct {
	AccessCount int64
}
