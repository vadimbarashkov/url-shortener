package postgres

import "github.com/jmoiron/sqlx"

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}
