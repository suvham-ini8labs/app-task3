package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"order-service/internal/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock for OrderServiceInterface
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID int, token string, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	args := m.Called(ctx, userID, token, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrderResponse), args.Error(1)
}

func (m *MockOrderService) GetOrder(ctx context.Context, id int, token string) (*models.OrderResponse, error) {
	args := m.Called(ctx, id, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrderResponse), args.Error(1)
}

func (m *MockOrderService) GetUserOrders(ctx context.Context, userID int, token string) ([]models.OrderResponse, error) {
	args := m.Called(ctx, userID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OrderResponse), args.Error(1)
}

func (m *MockOrderService) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockUserClient is a mock for UserServiceClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUser(ctx context.Context, id int, token string) (*models.UserInfo, error) {
	args := m.Called(ctx, id, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockUserClient) ValidateToken(ctx context.Context, token string) (*models.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockUserClient) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

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

func TestOrderHandlers_CreateOrder_Success(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	userID := 1
	token := "valid-token"
	reqBody := models.CreateOrderRequest{BookID: 101, Quantity: 2}
	expectedResponse := &models.OrderResponse{ID: 1, UserID: userID}

	mockUserClient.On("ValidateToken", mock.Anything, token).Return(&models.UserInfo{ID: userID}, nil)
	mockService.On("CreateOrder", mock.Anything, userID, token, &reqBody).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqBody)
	r := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.OrderResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, expectedResponse.ID, resp.ID)
}

func TestOrderHandlers_CreateOrder_Unauthorized(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	r := httptest.NewRequest("POST", "/orders", nil)
	// No Authorization header
	w := httptest.NewRecorder()

	h.CreateOrder(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandlers_GetOrder_Success(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	userID := 1
	orderID := 1
	token := "valid-token"
	expectedOrder := &models.OrderResponse{ID: orderID, UserID: userID}

	mockUserClient.On("ValidateToken", mock.Anything, token).Return(&models.UserInfo{ID: userID}, nil)
	mockService.On("GetOrder", mock.Anything, orderID, token).Return(expectedOrder, nil)

	r := httptest.NewRequest("GET", "/orders/1", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r = mux.SetURLVars(r, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetOrder(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.OrderResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, orderID, resp.ID)
}

func TestOrderHandlers_GetOrder_Forbidden(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	userID := 1
	otherUserID := 2
	orderID := 1
	token := "valid-token"
	expectedOrder := &models.OrderResponse{ID: orderID, UserID: otherUserID} // Order belongs to someone else

	mockUserClient.On("ValidateToken", mock.Anything, token).Return(&models.UserInfo{ID: userID}, nil)
	mockService.On("GetOrder", mock.Anything, orderID, token).Return(expectedOrder, nil)

	r := httptest.NewRequest("GET", "/orders/1", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r = mux.SetURLVars(r, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetOrder(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOrderHandlers_GetUserOrders_Success(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	userID := 1
	token := "valid-token"
	expectedOrders := []models.OrderResponse{{ID: 1, UserID: userID}}

	mockUserClient.On("ValidateToken", mock.Anything, token).Return(&models.UserInfo{ID: userID}, nil)
	mockService.On("GetUserOrders", mock.Anything, userID, token).Return(expectedOrders, nil)

	r := httptest.NewRequest("GET", "/users/1/orders", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r = mux.SetURLVars(r, map[string]string{"userId": "1"})
	w := httptest.NewRecorder()

	h.GetUserOrders(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandlers_Health_Healthy(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	mockService.On("Health", mock.Anything).Return(nil)

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandlers_Health_Unhealthy(t *testing.T) {
	mockService := new(MockOrderService)
	mockUserClient := new(MockUserClient)
	mockLogger := new(MockLogger)

	h := NewOrderHandlers(mockService, mockUserClient, mockLogger)

	mockService.On("Health", mock.Anything).Return(errors.New("unhealthy"))

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
