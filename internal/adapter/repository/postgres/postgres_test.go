package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/vadimbarashkov/url-shortener/internal/entity"
)

type URLRepositoryTestSuite struct {
	suite.Suite
	errUnknown      error
	errAffectedRows error
	columns         []string
	mock            sqlmock.Sqlmock
	repo            *URLRepository
}

func (suite *URLRepositoryTestSuite) SetupSuite() {
	suite.errUnknown = errors.New("unknown error")
	suite.errAffectedRows = errors.New("affected rows error")
	suite.columns = []string{"id", "short_code", "original_url", "access_count", "created_at", "updated_at"}
}

func (suite *URLRepositoryTestSuite) SetupSubTest() {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		suite.T().Fatalf("Failed to create mock database: %v", err)
	}
	suite.T().Cleanup(func() {
		mockDB.Close()
	})

	db := sqlx.NewDb(mockDB, "sqlmock")
	suite.T().Cleanup(func() {
		db.Close()
	})

	suite.mock = mock
	suite.repo = NewURLRepository(db)
}

func (suite *URLRepositoryTestSuite) TearDownSubTest() {
	suite.NoError(suite.mock.ExpectationsWereMet())
}

func (suite *URLRepositoryTestSuite) TestSave() {
	suite.Run("short code exists", func() {
		suite.mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("abc123", "https://example.com").
			WillReturnError(&pgconn.PgError{Code: uniqueViolationErrCode})

		url, err := suite.repo.Save(context.Background(), "abc123", "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, entity.ErrShortCodeExists)
		suite.Nil(url)
	})

	suite.Run("unknown error", func() {
		suite.mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("abc123", "https://example.com").
			WillReturnError(suite.errUnknown)

		url, err := suite.repo.Save(context.Background(), "abc123", "https://example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		rows := sqlmock.NewRows(suite.columns).
			AddRow(0, "abc123", "https://example.com", 0, time.Time{}, time.Time{})

		suite.mock.ExpectQuery(`INSERT INTO urls`).
			WithArgs("abc123", "https://example.com").
			WillReturnRows(rows)

		url, err := suite.repo.Save(context.Background(), "abc123", "https://example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) TestRetrieveByShortCode() {
	suite.Run("url not found", func() {
		suite.mock.ExpectQuery(`SELECT (.+) FROM urls`).
			WithArgs("abc123").
			WillReturnError(sql.ErrNoRows)

		url, err := suite.repo.RetrieveByShortCode(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, entity.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("unknown error", func() {
		suite.mock.ExpectQuery(`SELECT (.+) FROM urls`).
			WithArgs("abc123").
			WillReturnError(suite.errUnknown)

		url, err := suite.repo.RetrieveByShortCode(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		rows := sqlmock.NewRows(suite.columns).
			AddRow(0, "abc123", "https://example.com", 0, time.Time{}, time.Time{})

		suite.mock.ExpectQuery(`SELECT (.+) FROM urls`).
			WithArgs("abc123").
			WillReturnRows(rows)

		url, err := suite.repo.RetrieveByShortCode(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) TestRetrieveAndUpdateStats() {
	suite.Run("url not found", func() {
		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("abc123").
			WillReturnError(sql.ErrNoRows)

		url, err := suite.repo.RetrieveAndUpdateStats(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, entity.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("unknown error", func() {
		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("abc123").
			WillReturnError(suite.errUnknown)

		url, err := suite.repo.RetrieveAndUpdateStats(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		rows := sqlmock.NewRows(suite.columns).
			AddRow(0, "abc123", "https://example.com", 0, time.Time{}, time.Time{})

		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("abc123").
			WillReturnRows(rows)

		url, err := suite.repo.RetrieveAndUpdateStats(context.Background(), "abc123")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) TestUpdate() {
	suite.Run("url nof found", func() {
		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("https://new-example.com", "abc123").
			WillReturnError(sql.ErrNoRows)

		url, err := suite.repo.Update(context.Background(), "abc123", "https://new-example.com")

		suite.Error(err)
		suite.ErrorIs(err, entity.ErrURLNotFound)
		suite.Nil(url)
	})

	suite.Run("unknown error", func() {
		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("https://new-example.com", "abc123").
			WillReturnError(suite.errUnknown)

		url, err := suite.repo.Update(context.Background(), "abc123", "https://new-example.com")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
		suite.Nil(url)
	})

	suite.Run("success", func() {
		rows := sqlmock.NewRows(suite.columns).
			AddRow(0, "abc123", "https://new-example.com", 0, time.Time{}, time.Time{})

		suite.mock.ExpectQuery(`UPDATE urls`).
			WithArgs("https://new-example.com", "abc123").
			WillReturnRows(rows)

		url, err := suite.repo.Update(context.Background(), "abc123", "https://new-example.com")

		suite.NoError(err)
		suite.NotNil(url)
		suite.Equal("abc123", url.ShortCode)
		suite.Equal("https://new-example.com", url.OriginalURL)
		suite.Zero(url.AccessCount)
	})
}

func (suite *URLRepositoryTestSuite) TestRemove() {
	suite.Run("unknown error", func() {
		suite.mock.ExpectExec(`DELETE FROM urls`).
			WithArgs("abc123").
			WillReturnError(suite.errUnknown)

		err := suite.repo.Remove(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errUnknown)
	})

	suite.Run("rows affected error", func() {
		suite.mock.ExpectExec(`DELETE FROM urls`).
			WithArgs("abc123").
			WillReturnResult(sqlmock.NewErrorResult(suite.errAffectedRows))

		err := suite.repo.Remove(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, suite.errAffectedRows)
	})

	suite.Run("url not found", func() {
		suite.mock.ExpectExec(`DELETE FROM urls`).
			WithArgs("abc123").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := suite.repo.Remove(context.Background(), "abc123")

		suite.Error(err)
		suite.ErrorIs(err, entity.ErrURLNotFound)
	})

	suite.Run("success", func() {
		suite.mock.ExpectExec(`DELETE FROM urls`).
			WithArgs("abc123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := suite.repo.Remove(context.Background(), "abc123")

		suite.NoError(err)
	})
}

func TestURLRepository(t *testing.T) {
	suite.Run(t, new(URLRepositoryTestSuite))
}
