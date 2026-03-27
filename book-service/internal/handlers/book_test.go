package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"book-service/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBookService is a mock implementation of BookServiceInterface
type MockBookService struct {
	mock.Mock
}

func (m *MockBookService) CreateBook(ctx context.Context, req *models.CreateBookRequest) (*models.Book, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) GetBook(ctx context.Context, id int) (*models.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) ListBooks(ctx context.Context) ([]models.Book, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookService) UpdateBook(ctx context.Context, id int, req *models.UpdateBookRequest) (*models.Book, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) DeleteBook(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookService) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestBookHandlers_CreateBook(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("success", func(t *testing.T) {
		reqBody := models.CreateBookRequest{Title: "Test", Author: "Author"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		svc.On("CreateBook", mock.Anything, &reqBody).Return(&models.Book{ID: 1, Title: "Test"}, nil).Once()

		h.CreateBook(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var book models.Book
		json.NewDecoder(rr.Body).Decode(&book)
		assert.Equal(t, "Test", book.Title)
		svc.AssertExpectations(t)
	})

	t.Run("bad_request_body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/books", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()
		h.CreateBook(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service_error", func(t *testing.T) {
		reqBody := models.CreateBookRequest{Title: "Test"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		svc.On("CreateBook", mock.Anything, &reqBody).Return(nil, errors.New("service error")).Once()

		h.CreateBook(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestBookHandlers_GetBook(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/books/1", nil)
		rr := httptest.NewRecorder()

		// Setting up mux vars
		req = mux.SetURLVars(req, map[string]string{"id": "1"})

		svc.On("GetBook", mock.Anything, 1).Return(&models.Book{ID: 1, Title: "Test"}, nil).Once()

		h.GetBook(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var book models.Book
		json.NewDecoder(rr.Body).Decode(&book)
		assert.Equal(t, 1, book.ID)
		svc.AssertExpectations(t)
	})

	t.Run("invalid_id_format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/books/abc", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})

		h.GetBook(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/books/1", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": "1"})

		svc.On("GetBook", mock.Anything, 1).Return(nil, errors.New("book not found")).Once()

		h.GetBook(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestBookHandlers_ListBooks(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/books", nil)
		rr := httptest.NewRecorder()

		books := []models.Book{{ID: 1, Title: "Test"}}
		svc.On("ListBooks", mock.Anything).Return(books, nil).Once()

		h.ListBooks(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var res []models.Book
		json.NewDecoder(rr.Body).Decode(&res)
		assert.Len(t, res, 1)
		svc.AssertExpectations(t)
	})
}

func TestBookHandlers_UpdateBook(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("success", func(t *testing.T) {
		reqBody := models.UpdateBookRequest{Title: "New"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/books/1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": "1"})

		svc.On("UpdateBook", mock.Anything, 1, &reqBody).Return(&models.Book{ID: 1, Title: "New"}, nil).Once()

		h.UpdateBook(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestBookHandlers_DeleteBook(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/books/1", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"id": "1"})

		svc.On("DeleteBook", mock.Anything, 1).Return(nil).Once()

		h.DeleteBook(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestBookHandlers_Health(t *testing.T) {
	svc := new(MockBookService)
	h := NewBookHandlers(svc)

	t.Run("healthy", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		svc.On("Health", mock.Anything).Return(nil).Once()

		h.Health(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("unhealthy", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		svc.On("Health", mock.Anything).Return(errors.New("error")).Once()

		h.Health(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
		svc.AssertExpectations(t)
	})
}
