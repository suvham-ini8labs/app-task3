package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"book-service/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBookRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()
	book := &models.Book{
		Title:  "Test Book",
		Author: "Test Author",
		Price:  19.99,
		Stock:  10,
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now())
		
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Title, book.Author, book.Price, book.Stock).
			WillReturnRows(rows)

		err := repo.Create(ctx, book)
		assert.NoError(t, err)
		assert.Equal(t, 1, book.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db_error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO books").
			WillReturnError(fmt.Errorf("db error"))

		err := repo.Create(ctx, book)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create book")
	})
}

func TestBookRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Title", "Author", 10.0, 5, now, now)

		mock.ExpectQuery("SELECT (.+) FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		book, err := repo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, 1, book.ID)
		assert.Equal(t, "Title", book.Title)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM books").
			WithArgs(1).
			WillReturnError(sql.ErrNoRows)

		book, err := repo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Nil(t, book)
	})
}

func TestBookRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "B1", "A1", 10.0, 5, now, now).
			AddRow(2, "B2", "A2", 20.0, 10, now, now)

		mock.ExpectQuery("SELECT (.+) FROM books ORDER BY id DESC").
			WillReturnRows(rows)

		books, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, books, 2)
	})
}

func TestBookRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()
	book := &models.Book{Title: "Updated"}

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		mock.ExpectQuery("UPDATE books SET").
			WithArgs(book.Title, book.Author, book.Price, book.Stock, 1).
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		err := repo.Update(ctx, 1, book)
		assert.NoError(t, err)
		assert.Equal(t, 1, book.ID)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE books SET").
			WillReturnError(sql.ErrNoRows)

		err := repo.Update(ctx, 1, book)
		assert.Error(t, err)
		assert.Equal(t, "book not found", err.Error())
	})
}

func TestBookRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM books").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM books").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, 1)
		assert.Error(t, err)
		assert.Equal(t, "book not found", err.Error())
	})
}

func TestBookRepository_Health(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()

	t.Run("healthy", func(t *testing.T) {
		mock.ExpectPing()
		err := repo.Health(ctx)
		assert.NoError(t, err)
	})
}

func TestBookRepository_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &BookRepository{db: db}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT(.+) FROM books").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
		mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(100))
		mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"avg"}).AddRow(15.5))

		stats, err := repo.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 10, stats["total_books"])
		assert.Equal(t, 100, stats["total_stock"])
		assert.Equal(t, 15.5, stats["average_price"])
	})
}
