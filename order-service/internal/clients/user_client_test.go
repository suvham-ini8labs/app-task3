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

func TestUserClient_GetUser_Success(t *testing.T) {
	expectedUser := &models.UserInfo{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedUser)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewUserClient(server.URL, 5*time.Second, logger)

	user, err := client.GetUser(context.Background(), 1, "test-token")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUserClient_GetUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	logger := new(MockLogger)
	logger.On("Error", "User service returned error", mock.Anything).Return()
	client := NewUserClient(server.URL, 5*time.Second, logger)

	user, err := client.GetUser(context.Background(), 999, "test-token")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserClient_ValidateToken(t *testing.T) {
	logger := new(MockLogger)
	logger.On("Debug", "Validating token", mock.Anything).Return()
	client := NewUserClient("http://localhost:8081", 5*time.Second, logger)

	user, err := client.ValidateToken(context.Background(), "some-token")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID) // Based on the current mock implementation in user_client.go
	assert.Equal(t, "john", user.Username)
}

func TestUserClient_Health_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewUserClient(server.URL, 5*time.Second, logger)

	err := client.Health(context.Background())

	assert.NoError(t, err)
}

func TestUserClient_Health_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	logger := new(MockLogger)
	client := NewUserClient(server.URL, 5*time.Second, logger)

	err := client.Health(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user service unhealthy")
}
