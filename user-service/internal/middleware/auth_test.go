package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"user-service/internal/models"
	"user-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	return nil, nil
}
func (m *MockUserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	return nil, nil
}
func (m *MockUserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	return nil, nil
}
func (m *MockUserService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.User, error) {
	return nil, nil
}
func (m *MockUserService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserService) Health(ctx context.Context) error {
	return nil
}

func TestAuthMiddleware(t *testing.T) {
	mockService := new(MockUserService)
	log, _ := logger.New("debug", "console")
	authMiddleware := Auth(mockService, log)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, 1, userID)
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(nextHandler)

	t.Run("MissingHeader", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp models.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "Missing authorization header", resp.Error)
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp models.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "Invalid authorization header format", resp.Error)
	})

	t.Run("Success", func(t *testing.T) {
		user := &models.User{ID: 1, Username: "test", Email: "a@b.com"}
		mockService.On("ValidateToken", mock.Anything, "valid-token").Return(user, nil).Once()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
