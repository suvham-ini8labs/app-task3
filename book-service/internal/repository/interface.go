package repository

import (
	"context"

	"book-service/internal/models"
)

type BookRepositoryInterface interface {
	Create(ctx context.Context, book *models.Book) error
	GetByID(ctx context.Context, id int) (*models.Book, error)
	GetAll(ctx context.Context) ([]models.Book, error)
	Update(ctx context.Context, id int, book *models.Book) error
	Delete(ctx context.Context, id int) error
	Health(ctx context.Context) error
	Close() error
}
