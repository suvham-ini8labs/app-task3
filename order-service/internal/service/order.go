package service

import (
	"context"
	"fmt"

	"order-service/internal/clients"
	"order-service/internal/models"
	"order-service/internal/repository"
)

type OrderService struct {
	repo       repository.OrderRepositoryInterface
	bookClient clients.BookServiceClient
	userClient clients.UserServiceClient
	logger     Logger
}

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

var _ OrderServiceInterface = (*OrderService)(nil)

func NewOrderService(
	repo repository.OrderRepositoryInterface,
	bookClient clients.BookServiceClient,
	userClient clients.UserServiceClient,
	logger Logger,
) *OrderService {
	return &OrderService{
		repo:       repo,
		bookClient: bookClient,
		userClient: userClient,
		logger:     logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int, token string, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// Validate user exists with token
	user, err := s.userClient.GetUser(ctx, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Validate book exists and get details
	book, err := s.bookClient.GetBook(ctx, req.BookID)
	if err != nil {
		s.logger.Error("Failed to validate book", "book_id", req.BookID, "error", err)
		return nil, fmt.Errorf("book validation failed: %w", err)
	}

	// Check stock availability
	if book.Stock < req.Quantity {
		return nil, fmt.Errorf("insufficient stock: available %d, requested %d", book.Stock, req.Quantity)
	}

	// Calculate total price
	totalPrice := book.Price * float64(req.Quantity)

	// Create order
	order := &models.Order{
		UserID:     userID,
		BookID:     req.BookID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
		Status:     models.StatusConfirmed,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		s.logger.Error("Failed to create order", "error", err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Prepare response
	response := &models.OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		BookID:     order.BookID,
		Quantity:   order.Quantity,
		TotalPrice: order.TotalPrice,
		Status:     order.Status,
		Book:       book,
		User:       user,
		CreatedAt:  order.CreatedAt,
	}

	s.logger.Info("Order created successfully", "order_id", order.ID, "user_id", userID)
	return response, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id int, token string) (*models.OrderResponse, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	// Fetch book details
	book, err := s.bookClient.GetBook(ctx, order.BookID)
	if err != nil {
		s.logger.Error("Failed to fetch book details", "book_id", order.BookID, "error", err)
		// Continue without book details
	}

	// Fetch user details with token
	user, err := s.userClient.GetUser(ctx, order.UserID, token)
	if err != nil {
		s.logger.Error("Failed to fetch user details", "user_id", order.UserID, "error", err)
		// Continue without user details
	}

	response := &models.OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		BookID:     order.BookID,
		Quantity:   order.Quantity,
		TotalPrice: order.TotalPrice,
		Status:     order.Status,
		Book:       book,
		User:       user,
		CreatedAt:  order.CreatedAt,
	}

	return response, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int, token string) ([]models.OrderResponse, error) {
	// Validate user exists with token
	_, err := s.userClient.GetUser(ctx, userID, token)
	if err != nil {
		s.logger.Error("Failed to validate user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	orders, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build responses
	responses := make([]models.OrderResponse, 0, len(orders))
	for _, order := range orders {
		// Fetch book details
		book, _ := s.bookClient.GetBook(ctx, order.BookID)
		
		responses = append(responses, models.OrderResponse{
			ID:         order.ID,
			UserID:     order.UserID,
			BookID:     order.BookID,
			Quantity:   order.Quantity,
			TotalPrice: order.TotalPrice,
			Status:     order.Status,
			Book:       book,
			CreatedAt:  order.CreatedAt,
		})
	}

	return responses, nil
}

func (s *OrderService) Health(ctx context.Context) error {
	if err := s.repo.Health(ctx); err != nil {
		return err
	}
	if err := s.bookClient.Health(ctx); err != nil {
		return err
	}
	if err := s.userClient.Health(ctx); err != nil {
		return err
	}
	return nil
}
