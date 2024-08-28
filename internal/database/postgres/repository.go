package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

type urlRecord struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (r *urlRecord) ToURL() *models.URL {
	return &models.URL{
		ID:          r.ID,
		ShortCode:   r.ShortCode,
		OriginalURL: r.OriginalURL,
		AccessCount: r.AccessCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

func (r *URLRepository) Create(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.Create"

	rec := new(urlRecord)
	query := `INSERT INTO urls(short_code, original_url)
		VALUES ($1, $2)
		RETURNING *`

	err := r.db.GetContext(ctx, rec, query, shortCode, originalURL)
	if err != nil {
		if isUniqueViolationError(err) {
			return nil, fmt.Errorf("%s: %w", op, database.ErrShortCodeExists)
		}

		return nil, fmt.Errorf("%s: failed to create url record: %w", op, err)
	}

	return rec.ToURL(), nil
}

func (r *URLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.GetByShortCode"

	rec := new(urlRecord)
	query := `UPDATE urls
		SET access_count = access_count + 1
		WHERE short_code = $1
		RETURNING *`

	err := r.db.GetContext(ctx, rec, query, shortCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get url record: %w", op, err)
	}

	return rec.ToURL(), nil
}

func (r *URLRepository) Update(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.Update"

	rec := new(urlRecord)
	query := `UPDATE urls
		SET original_url = $1
		WHERE short_code = $2
		RETURNING *`

	err := r.db.GetContext(ctx, rec, query, originalURL, shortCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to update url record: %w", op, err)
	}

	return rec.ToURL(), nil
}
