package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

// urlRecord represetns the internal structure of a URL record stored in the database.
// It is used internally within the repository to map database columns to Go struct fields.
type urlRecord struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// toURL converts a urlRecord struct to a models.URL struct. This is used to transform
// data from the database format to the business layer format.
func (r *urlRecord) toURL() *models.URL {
	return &models.URL{
		ID:          r.ID,
		ShortCode:   r.ShortCode,
		OriginalURL: r.OriginalURL,
		AccessCount: r.AccessCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// URLRepository provides methods for interacting with URL records in the database.
// It is responsible for CRUD operations on the 'urls' table.
type URLRepository struct {
	db *sqlx.DB
}

// NewURLRepository creates a new instance of URLRepository. It requires a sqlx.DB object
// representing the database connection.
func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

// Create inserts a new URL record into the database with the specified short code and original url.
// If the short code already exists, it returns a database.ErrShortCodeExists error. On success, it
// returns the newly created models.URL object.
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

	return rec.toURL(), nil
}

// GetByShortCode retrieves a URL record from the database by its short code. It increments the
// access_count value for the record and returns the corresponding models.URL object. If no
// record is found, it returns a database.ErrURLNotFound error.
func (r *URLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.GetByShortCode"

	rec := new(urlRecord)
	query := `UPDATE urls
		SET access_count = access_count + 1
		WHERE short_code = $1
		RETURNING *`

	err := r.db.GetContext(ctx, rec, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get url record: %w", op, err)
	}

	return rec.toURL(), nil
}

// Update modifies the URL associated with the given short code. If the alias doesn't exist, it
// returns a database.ErrURLNotFound error. On success, it returns the updated models.URL object.
func (r *URLRepository) Update(ctx context.Context, shortCode, originalURL string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.Update"

	rec := new(urlRecord)
	query := `UPDATE urls
		SET original_url = $1
		WHERE short_code = $2
		RETURNING *`

	err := r.db.GetContext(ctx, rec, query, originalURL, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to update url record: %w", op, err)
	}

	return rec.toURL(), nil
}

// Delete removes a URL record from the database by its short code. If the short code doesn't
// exist, it returns a database.ErrURLNotFound error. On success, it returns nil.
func (r *URLRepository) Delete(ctx context.Context, shortCode string) error {
	const op = "database.postgres.URLRepository.Delete"

	query := `DELETE FROM urls
		WHERE short_code = $1`

	res, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to delete url record: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get number of affected rows: %w", op, err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
	}

	return nil
}

// GetStats retrieves a URL record from the database by its short code. It doesn't change the
// URL record, but only returns the corresponding models.URL object. If no record is found, it
// returns a database.ErrURLNotFound error.
func (r *URLRepository) GetStats(ctx context.Context, shortCode string) (*models.URL, error) {
	const op = "database.postgres.URLRepository.GetStats"

	rec := new(urlRecord)
	query := `SELECT * FROM urls
		WHERE short_code = $1`

	err := r.db.GetContext(ctx, rec, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, database.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get url record: %w", op, err)
	}

	return rec.toURL(), nil
}
