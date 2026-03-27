package service

import (
	"context"
	"errors"
	"testing"

	"book-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBookRepository is a mock implementation of BookRepositoryInterface
type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) Create(ctx context.Context, book *models.Book) error {
	args := m.Called(ctx, book)
	return args.Error(0)
}

func (m *MockBookRepository) GetByID(ctx context.Context, id int) (*models.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookRepository) GetAll(ctx context.Context) ([]models.Book, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookRepository) Update(ctx context.Context, id int, book *models.Book) error {
	args := m.Called(ctx, id, book)
	return args.Error(0)
}

func (m *MockBookRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookRepository) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBookRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestBookService_CreateBook(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := &models.CreateBookRequest{
			Title:  "Test Book",
			Author: "Test Author",
			Price:  10.99,
			Stock:  100,
		}
		repo.On("Create", ctx, mock.MatchedBy(func(b *models.Book) bool {
			return b.Title == req.Title && b.Author == req.Author
		})).Return(nil).Once()

		book, err := svc.CreateBook(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, req.Title, book.Title)
		repo.AssertExpectations(t)
	})

	t.Run("validation_error_title", func(t *testing.T) {
		req := &models.CreateBookRequest{
			Title: "",
		}
		book, err := svc.CreateBook(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("repo_error", func(t *testing.T) {
		req := &models.CreateBookRequest{
			Title:  "Test Book",
			Author: "Test Author",
			Price:  10.99,
			Stock:  100,
		}
		repo.On("Create", ctx, mock.Anything).Return(errors.New("db error")).Once()

		book, err := svc.CreateBook(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Contains(t, err.Error(), "failed to create book")
		repo.AssertExpectations(t)
	})
}

func TestBookService_GetBook(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", ctx, 1).Return(&models.Book{ID: 1, Title: "Test"}, nil).Once()
		book, err := svc.GetBook(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, book.ID)
		repo.AssertExpectations(t)
	})

	t.Run("invalid_id", func(t *testing.T) {
		book, err := svc.GetBook(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Equal(t, "invalid book id", err.Error())
	})

	t.Run("not_found", func(t *testing.T) {
		repo.On("GetByID", ctx, 1).Return(nil, nil).Once()
		book, err := svc.GetBook(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Equal(t, "book not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestBookService_ListBooks(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		books := []models.Book{{Title: "B1"}, {Title: "B2"}}
		repo.On("GetAll", ctx).Return(books, nil).Once()
		res, err := svc.ListBooks(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		repo.AssertExpectations(t)
	})

	t.Run("repo_error", func(t *testing.T) {
		repo.On("GetAll", ctx).Return(nil, errors.New("err")).Once()
		res, err := svc.ListBooks(ctx)
		assert.Error(t, err)
		assert.Nil(t, res)
		repo.AssertExpectations(t)
	})
}

func TestBookService_UpdateBook(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		existing := &models.Book{ID: 1, Title: "Old", Author: "Author"}
		req := &models.UpdateBookRequest{Title: "New"}
		repo.On("GetByID", ctx, 1).Return(existing, nil).Once()
		repo.On("Update", ctx, 1, mock.Anything).Return(nil).Once()

		book, err := svc.UpdateBook(ctx, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "New", book.Title)
		repo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		repo.On("GetByID", ctx, 1).Return(nil, nil).Once()
		book, err := svc.UpdateBook(ctx, 1, &models.UpdateBookRequest{})
		assert.Error(t, err)
		assert.Nil(t, book)
		assert.Equal(t, "book not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestBookService_DeleteBook(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.On("Delete", ctx, 1).Return(nil).Once()
		err := svc.DeleteBook(ctx, 1)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid_id", func(t *testing.T) {
		err := svc.DeleteBook(ctx, 0)
		assert.Error(t, err)
		assert.Equal(t, "invalid book id", err.Error())
	})
}

func TestBookService_Health(t *testing.T) {
	repo := new(MockBookRepository)
	svc := NewBookService(repo)
	ctx := context.Background()

	t.Run("healthy", func(t *testing.T) {
		repo.On("Health", ctx).Return(nil).Once()
		err := svc.Health(ctx)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("unhealthy", func(t *testing.T) {
		repo.On("Health", ctx).Return(errors.New("down")).Once()
		err := svc.Health(ctx)
		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}
