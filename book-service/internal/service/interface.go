package service

import (
	"context"

	"book-service/internal/models"
)

type BookServiceInterface interface {
	CreateBook(ctx context.Context, req *models.CreateBookRequest) (*models.Book, error)
	GetBook(ctx context.Context, id int) (*models.Book, error)
	ListBooks(ctx context.Context) ([]models.Book, error)
	UpdateBook(ctx context.Context, id int, req *models.UpdateBookRequest) (*models.Book, error)
	DeleteBook(ctx context.Context, id int) error
	Health(ctx context.Context) error
}
