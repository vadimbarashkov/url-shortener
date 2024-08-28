package database

import "errors"

var (
	ErrShortCodeExists = errors.New("short code exists")
	ErrURLNotFound     = errors.New("url not found")
)
