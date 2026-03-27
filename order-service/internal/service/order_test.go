package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"order-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock for OrderRepositoryInterface
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id int) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id int, status models.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockBookClient is a mock for BookServiceClient
type MockBookClient struct {
	mock.Mock
}

func (m *MockBookClient) GetBook(ctx context.Context, id int) (*models.BookInfo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BookInfo), args.Error(1)
}

func (m *MockBookClient) Health(ctx context.Context) error {
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

func TestCreateOrder_Success(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	req := &models.CreateOrderRequest{
		BookID:   101,
		Quantity: 2,
	}

	user := &models.UserInfo{ID: userID, Username: "testuser"}
	book := &models.BookInfo{ID: 101, Title: "Test Book", Stock: 10, Price: 25.0}

	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	bookClient.On("GetBook", ctx, req.BookID).Return(book, nil)
	repo.On("Create", ctx, mock.MatchedBy(func(order *models.Order) bool {
		return order.UserID == userID && order.BookID == req.BookID && order.Quantity == req.Quantity && order.TotalPrice == 50.0
	})).Return(nil)
	logger.On("Info", "Order created successfully", mock.Anything).Return()

	resp, err := s.CreateOrder(ctx, userID, token, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, 50.0, resp.TotalPrice)
	assert.Equal(t, user, resp.User)
	assert.Equal(t, book, resp.Book)
	repo.AssertExpectations(t)
	bookClient.AssertExpectations(t)
	userClient.AssertExpectations(t)
}

func TestCreateOrder_UserValidationFailed(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "invalid-token"
	req := &models.CreateOrderRequest{BookID: 101, Quantity: 2}

	userClient.On("GetUser", ctx, userID, token).Return(nil, errors.New("invalid token"))
	logger.On("Error", "Failed to validate user", mock.Anything).Return()

	resp, err := s.CreateOrder(ctx, userID, token, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "user validation failed")
}

func TestCreateOrder_BookValidationFailed(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	req := &models.CreateOrderRequest{BookID: 101, Quantity: 2}

	user := &models.UserInfo{ID: userID}
	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	bookClient.On("GetBook", ctx, req.BookID).Return(nil, errors.New("book not found"))
	logger.On("Error", "Failed to validate book", mock.Anything).Return()

	resp, err := s.CreateOrder(ctx, userID, token, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "book validation failed")
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	req := &models.CreateOrderRequest{BookID: 101, Quantity: 5}

	user := &models.UserInfo{ID: userID}
	book := &models.BookInfo{ID: 101, Stock: 2}

	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	bookClient.On("GetBook", ctx, req.BookID).Return(book, nil)

	resp, err := s.CreateOrder(ctx, userID, token, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "insufficient stock")
}

func TestGetOrder_Success(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	orderID := 1
	token := "valid-token"
	order := &models.Order{ID: orderID, UserID: 1, BookID: 101, Quantity: 1, TotalPrice: 20.0, CreatedAt: time.Now()}
	user := &models.UserInfo{ID: 1}
	book := &models.BookInfo{ID: 101}

	repo.On("GetByID", ctx, orderID).Return(order, nil)
	bookClient.On("GetBook", ctx, order.BookID).Return(book, nil)
	userClient.On("GetUser", ctx, order.UserID, token).Return(user, nil)

	resp, err := s.GetOrder(ctx, orderID, token)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, orderID, resp.ID)
	assert.Equal(t, user, resp.User)
	assert.Equal(t, book, resp.Book)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	orderID := 999

	repo.On("GetByID", ctx, orderID).Return(nil, nil)

	resp, err := s.GetOrder(ctx, orderID, "token")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "order not found")
}

func TestGetUserOrders_Success(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	orders := []models.Order{
		{ID: 1, UserID: userID, BookID: 101},
		{ID: 2, UserID: userID, BookID: 102},
	}
	user := &models.UserInfo{ID: userID}

	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	repo.On("GetByUserID", ctx, userID).Return(orders, nil)
	bookClient.On("GetBook", ctx, 101).Return(&models.BookInfo{ID: 101}, nil)
	bookClient.On("GetBook", ctx, 102).Return(&models.BookInfo{ID: 102}, nil)

	resp, err := s.GetUserOrders(ctx, userID, token)

	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestHealth_Success(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	repo.On("Health", ctx).Return(nil)
	bookClient.On("Health", ctx).Return(nil)
	userClient.On("Health", ctx).Return(nil)

	err := s.Health(ctx)

	assert.NoError(t, err)
}

func TestHealth_RepoUnhealthy(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	repo.On("Health", ctx).Return(errors.New("db error"))

	err := s.Health(ctx)

	assert.Error(t, err)
}

func TestCreateOrder_RepoCreateFailed(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	req := &models.CreateOrderRequest{BookID: 101, Quantity: 2}

	user := &models.UserInfo{ID: userID}
	book := &models.BookInfo{ID: 101, Stock: 10, Price: 25.0}

	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	bookClient.On("GetBook", ctx, req.BookID).Return(book, nil)
	repo.On("Create", ctx, mock.Anything).Return(errors.New("db insert fail"))
	logger.On("Error", "Failed to create order", mock.Anything).Return()

	resp, err := s.CreateOrder(ctx, userID, token, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create order")
}

func TestGetOrder_RepoError(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	orderID := 1

	repo.On("GetByID", ctx, orderID).Return(nil, errors.New("db fetch error"))

	resp, err := s.GetOrder(ctx, orderID, "token")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "db fetch error")
}

func TestGetOrder_PartialDetails(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	orderID := 1
	token := "valid-token"
	order := &models.Order{ID: orderID, UserID: 1, BookID: 101, Quantity: 1, TotalPrice: 20.0}

	repo.On("GetByID", ctx, orderID).Return(order, nil)
	// Partial failure: book details fail
	bookClient.On("GetBook", ctx, order.BookID).Return(nil, errors.New("book component down"))
	// Partial failure: user details fail
	userClient.On("GetUser", ctx, order.UserID, token).Return(nil, errors.New("user component down"))

	logger.On("Error", "Failed to fetch book details", mock.Anything).Return()
	logger.On("Error", "Failed to fetch user details", mock.Anything).Return()

	resp, err := s.GetOrder(ctx, orderID, token)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, orderID, resp.ID)
	assert.Nil(t, resp.Book)
	assert.Nil(t, resp.User)
}

func TestGetUserOrders_RepoError(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	userID := 1
	token := "valid-token"
	user := &models.UserInfo{ID: userID}

	userClient.On("GetUser", ctx, userID, token).Return(user, nil)
	repo.On("GetByUserID", ctx, userID).Return(nil, errors.New("db list error"))

	resp, err := s.GetUserOrders(ctx, userID, token)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "db list error")
}

func TestHealth_BookClientUnhealthy(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	repo.On("Health", ctx).Return(nil)
	bookClient.On("Health", ctx).Return(errors.New("book service down"))

	err := s.Health(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "book service down")
}

func TestHealth_UserClientUnhealthy(t *testing.T) {
	repo := new(MockOrderRepository)
	bookClient := new(MockBookClient)
	userClient := new(MockUserClient)
	logger := new(MockLogger)

	s := NewOrderService(repo, bookClient, userClient, logger)

	ctx := context.Background()
	repo.On("Health", ctx).Return(nil)
	bookClient.On("Health", ctx).Return(nil)
	userClient.On("Health", ctx).Return(errors.New("user service down"))

	err := s.Health(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user service down")
}
