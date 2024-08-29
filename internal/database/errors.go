package database

import "errors"

var (
	// ErrShortCodeExists is returned when an attempt is made to create
	// a new shortened URL with a short code that already exists.
	ErrShortCodeExists = errors.New("short code exists")
	// ErrURLNotFound is returned when an attempt is made to retrieve
	// a URL using a short code that doesn't exist.
	ErrURLNotFound = errors.New("url not found")
)
