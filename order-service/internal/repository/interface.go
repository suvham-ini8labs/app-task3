package repository

import (
	"context"

	"order-service/internal/models"
)

type OrderRepositoryInterface interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id int) (*models.Order, error)
	GetByUserID(ctx context.Context, userID int) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id int, status models.OrderStatus) error
	Health(ctx context.Context) error
	Close() error
}
