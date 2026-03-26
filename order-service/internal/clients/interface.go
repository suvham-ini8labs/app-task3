package clients

import (
	"context"

	"order-service/internal/models"
)

type BookServiceClient interface {
	GetBook(ctx context.Context, id int) (*models.BookInfo, error)
	Health(ctx context.Context) error
}

type UserServiceClient interface {
	GetUser(ctx context.Context, id int, token string) (*models.UserInfo, error)
	ValidateToken(ctx context.Context, token string) (*models.UserInfo, error)
	Health(ctx context.Context) error
}
