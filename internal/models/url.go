package models

import "time"

type URL struct {
	ID          int64
	ShortCode   string
	OriginalURL string
	AccessCount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
