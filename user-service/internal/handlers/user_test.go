package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"user-service/internal/middleware"
	"user-service/internal/models"
	"user-service/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestUserHandlers(t *testing.T) {
	mockService := new(MockUserService)
	log, _ := logger.New("info", "console")
	h := NewUserHandlers(mockService, log)

	t.Run("Register_Success", func(t *testing.T) {
		req := &models.RegisterRequest{Username: "test", Email: "a@b.com", Password: "pwd"}
		user := &models.User{ID: 1, Username: "test", Email: "a@b.com"}

		mockService.On("Register", mock.Anything, req).Return(user, nil).Once()

		body, _ := json.Marshal(req)
		r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Register(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp models.User
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, user.ID, resp.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("Register_Conflict", func(t *testing.T) {
		req := &models.RegisterRequest{Email: "a@b.com"}
		mockService.On("Register", mock.Anything, req).Return(nil, errors.New("email already registered")).Once()

		body, _ := json.Marshal(req)
		r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Register(w, r)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Login_Success", func(t *testing.T) {
		req := &models.LoginRequest{Email: "a@b.com", Password: "pwd"}
		resp := &models.LoginResponse{Token: "test-token"}

		mockService.On("Login", mock.Anything, req).Return(resp, nil).Once()

		body, _ := json.Marshal(req)
		r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.Login(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		var loginResp models.LoginResponse
		json.NewDecoder(w.Body).Decode(&loginResp)
		assert.Equal(t, "test-token", loginResp.Token)
		mockService.AssertExpectations(t)
	})

	t.Run("GetUser_Success", func(t *testing.T) {
		id := 1
		user := &models.User{ID: id, Username: "test"}

		mockService.On("GetUser", mock.Anything, id).Return(user, nil).Once()

		r := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(id)})
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, id)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUser(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.User
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, id, resp.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("GetUser_Forbidden", func(t *testing.T) {
		id := 1
		r := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(id), nil)
		r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(id)})
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 2) // different ID
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUser(w, r)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Health_Success", func(t *testing.T) {
		mockService.On("Health", mock.Anything).Return(nil).Once()

		r := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		h.Health(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
