package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/url-shortener/internal/database"
	"github.com/vadimbarashkov/url-shortener/internal/models"
)

var errUnknown = errors.New("unknown error")

var columns = []string{"id", "short_code", "original_url", "access_count", "created_at", "updated_at"}

func setupURLRepository(t testing.TB) (*URLRepository, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewURLRepository(db)

	t.Cleanup(func() {
		mockDB.Close()
		db.Close()
	})

	return repo, mock
}

func TestURLRepository_Create(t *testing.T) {
	t.Run("short code exists", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("code1", "https://example.com").
			WillReturnError(&pgconn.PgError{Code: uniqueViolationErrCode})

		url, err := repo.Create(context.TODO(), "code1", "https://example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, database.ErrShortCodeExists)
		assert.Nil(t, url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unknown error", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("code1", "https://example.com").
			WillReturnError(errUnknown)

		url, err := repo.Create(context.TODO(), "code1", "https://example.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(0, "code1", "https://example.com", 0, time.Time{}, time.Time{})

		mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("code1", "https://example.com").
			WillReturnRows(rows)

		wantURL := models.URL{
			ShortCode:   "code1",
			OriginalURL: "https://example.com",
		}

		url, err := repo.Create(context.TODO(), "code1", "https://example.com")

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, wantURL, *url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestURLRepository_GetByShortCode(t *testing.T) {
	t.Run("url not found", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		mock.ExpectQuery(`UPDATE urls`).
			WithArgs("code2").
			WillReturnError(sql.ErrNoRows)

		url, err := repo.GetByShortCode(context.TODO(), "code2")

		assert.Error(t, err)
		assert.ErrorIs(t, err, database.ErrURLNotFound)
		assert.Nil(t, url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unknown error", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		mock.ExpectQuery(`UPDATE urls`).
			WithArgs("code1").
			WillReturnError(errUnknown)

		url, err := repo.GetByShortCode(context.TODO(), "code1")

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := setupURLRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(0, "code1", "https://example.com", 1, time.Time{}, time.Time{})

		mock.ExpectQuery(`UPDATE urls`).
			WithArgs("code1").
			WillReturnRows(rows)

		wantURL := models.URL{
			ShortCode:   "code1",
			OriginalURL: "https://example.com",
			AccessCount: 1,
		}

		url, err := repo.GetByShortCode(context.TODO(), "code1")

		assert.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, wantURL, *url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
