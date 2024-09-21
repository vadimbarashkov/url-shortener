package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

const uniqueViolationErrCode = "23505"

func isUniqueViolationError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.SQLState() == uniqueViolationErrCode
}

type urlDB struct {
	ID          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalURL string    `db:"original_url"`
	AccessCount int64     `db:"access_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (u *urlDB) toEntity() *entity.URL {
	return &entity.URL{
		ID:          u.ID,
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalURL,
		URLStats: entity.URLStats{
			AccessCount: u.AccessCount,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Save(ctx context.Context, shortCode, originalURL string) (*entity.URL, error) {
	const op = "adapter.repository.postgres.URLRepository.Save"
	const query = `INSERT INTO urls(short_code, original_url) VALUES ($1, $2) RETURNING *`

	var url urlDB

	if err := r.db.GetContext(ctx, &url, query, shortCode, originalURL); err != nil {
		if isUniqueViolationError(err) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrShortCodeExists)
		}

		return nil, fmt.Errorf("%s: failed to insert into urls table: %w", op, err)
	}

	return url.toEntity(), nil
}

func (r *URLRepository) RetrieveByShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "adapter.repository.postgres.URLRepository.RetrieveByShortCode"
	const query = `SELECT * FROM urls WHERE short_code = $1`

	var url urlDB

	if err := r.db.GetContext(ctx, &url, query, shortCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get row from urls table: %w", op, err)
	}

	return url.toEntity(), nil
}

func (r *URLRepository) RetrieveAndUpdateStats(ctx context.Context, shortCode string) (*entity.URL, error) {
	const op = "adapter.repository.postgres.URLRepository.RetrieveAndUpdateStats"
	const query = `UPDATE urls SET access_count = access_count + 1 WHERE short_code = $1 RETURNING *`

	var url urlDB

	if err := r.db.GetContext(ctx, &url, query, shortCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get and update urls table row: %w", op, err)
	}

	return url.toEntity(), nil
}

func (r *URLRepository) Update(ctx context.Context, shortCode, originalURL string) (*entity.URL, error) {
	const op = "adapter.repository.postgres.URLRepository.Update"
	const query = `UPDATE urls SET original_url = $1 WHERE short_code = $2 RETURNING *`

	var url urlDB

	if err := r.db.GetContext(ctx, &url, query, originalURL, shortCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrURLNotFound)
		}

		return nil, fmt.Errorf("%s: failed to update urls table row: %w", op, err)
	}

	return url.toEntity(), nil
}

func (r *URLRepository) Remove(ctx context.Context, shortCode string) error {
	const op = "adapter.repository.postgres.URLRepository.Remove"
	const query = `DELETE FROM urls WHERE short_code = $1`

	res, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("%s: failed to delete from urls table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get number of affected rows: %w", op, err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("%s: %w", op, entity.ErrURLNotFound)
	}

	return nil
}
