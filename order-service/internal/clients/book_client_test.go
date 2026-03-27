package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock for Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func TestBookClient_GetBook_Success(t *testing.T) {
	expectedBook := &models.BookInfo{
		ID:    101,
		Title: "Test Book",
		Price: 29.99,
		Stock: 5,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/books/101", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedBook)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewBookClient(server.URL, 5*time.Second, logger)

	book, err := client.GetBook(context.Background(), 101)

	assert.NoError(t, err)
	assert.Equal(t, expectedBook, book)
}

func TestBookClient_GetBook_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	logger := new(MockLogger)
	logger.On("Error", "Book service returned error", mock.Anything).Return()
	client := NewBookClient(server.URL, 5*time.Second, logger)

	book, err := client.GetBook(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, book)
	assert.Contains(t, err.Error(), "book not found")
}

func TestBookClient_Health_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewBookClient(server.URL, 5*time.Second, logger)

	err := client.Health(context.Background())

	assert.NoError(t, err)
}

func TestBookClient_Health_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewBookClient(server.URL, 5*time.Second, logger)

	err := client.Health(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "book service unhealthy")
}
