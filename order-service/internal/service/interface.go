package service

import (
	"context"

	"order-service/internal/models"
)

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID int, token string, req *models.CreateOrderRequest) (*models.OrderResponse, error)
	GetOrder(ctx context.Context, id int, token string) (*models.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID int, token string) ([]models.OrderResponse, error)
	Health(ctx context.Context) error
}
