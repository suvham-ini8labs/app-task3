package service

import (
	"context"
	"user-service/internal/models"
)

type UserService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.User, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	Health(ctx context.Context) error
}
