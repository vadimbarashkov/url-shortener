package models

import "time"

// URL represents a shortened URL and its associated metadata.
type URL struct {
	// ID is the unique identifier for the shortened URL record.
	ID int64
	// ShortCode it the short code or key associated with the original URL.
	ShortCode string
	// OriginalURL is the original, full-length URL that the short code points to.
	OriginalURL string
	// AccessCount tracks the number of times the shortened URL has been accessed.
	AccessCount int64
	// CreatedAt is the timestamp indicating when the shortened URL was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp indicating when the shortened URL was last updated.
	UpdatedAt time.Time
}
