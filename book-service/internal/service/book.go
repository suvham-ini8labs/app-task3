package service

import (
	"context"
	"errors"
	"fmt"

	"book-service/internal/models"
	"book-service/internal/repository"
)

type BookService struct {
	repo repository.BookRepositoryInterface
}

// Ensure BookService implements BookServiceInterface
var _ BookServiceInterface = (*BookService)(nil)

func NewBookService(repo repository.BookRepositoryInterface) *BookService {
	return &BookService{
		repo: repo,
	}
}

func (s *BookService) CreateBook(ctx context.Context, req *models.CreateBookRequest) (*models.Book, error) {
	// Validate input
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.Author == "" {
		return nil, errors.New("author is required")
	}
	if req.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}
	if req.Stock < 0 {
		return nil, errors.New("stock must be non-negative")
	}

	book := &models.Book{
		Title:  req.Title,
		Author: req.Author,
		Price:  req.Price,
		Stock:  req.Stock,
	}

	if err := s.repo.Create(ctx, book); err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	return book, nil
}

func (s *BookService) GetBook(ctx context.Context, id int) (*models.Book, error) {
	if id <= 0 {
		return nil, errors.New("invalid book id")
	}

	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("book not found")
	}

	return book, nil
}

func (s *BookService) ListBooks(ctx context.Context) ([]models.Book, error) {
	books, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list books: %w", err)
	}
	return books, nil
}

func (s *BookService) UpdateBook(ctx context.Context, id int, req *models.UpdateBookRequest) (*models.Book, error) {
	if id <= 0 {
		return nil, errors.New("invalid book id")
	}

	// Get existing book
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("book not found")
	}

	// Update fields if provided
	if req.Title != "" {
		book.Title = req.Title
	}
	if req.Author != "" {
		book.Author = req.Author
	}
	if req.Price >= 0 {
		book.Price = req.Price
	}
	if req.Stock >= 0 {
		book.Stock = req.Stock
	}

	// Validate updated book
	if book.Title == "" {
		return nil, errors.New("title cannot be empty")
	}
	if book.Author == "" {
		return nil, errors.New("author cannot be empty")
	}

	if err := s.repo.Update(ctx, id, book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) DeleteBook(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid book id")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *BookService) Health(ctx context.Context) error {
	return s.repo.Health(ctx)
}
