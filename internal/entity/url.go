// Package entity defines the entities and errors used in the application.
// It includes the URL struct, which represents a shortened URL, along with its
// associated metadata, and any relevant error definitions.
package entity

import (
	"errors"
	"time"
)

var (
	// ErrShortCodeExists is returned when attempting to create a URL with a short code that already exists.
	ErrShortCodeExists = errors.New("short code exists")
	// ErrURLNotFound is returned when a URL with the specified short code cannot be found.
	ErrURLNotFound = errors.New("url not found")
)

// URL represents a shortened URL.
type URL struct {
	ID          int64     // ID is the unique identifier of the URL in the database.
	ShortCode   string    // ShortCode is the generated code used to shorten the original URL.
	OriginalURL string    // OriginalURL is the full URL that the short code resolves to.
	URLStats              // URLStats contains statistics about the URL.
	CreatedAt   time.Time // CreatedAt is the timestamp when the URL was created.
	UpdatedAt   time.Time // UpdatedAt is the timestamp when the URL was last updated.
}

// URLStats contains statistics related to a shortened URL.
type URLStats struct {
	AccessCount int64 // AccessCount is the number of times the shortened URL has been accessed.
}
