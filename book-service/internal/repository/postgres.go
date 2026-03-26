package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "book-service/internal/models"
    "book-service/internal/config"

    _ "github.com/lib/pq"
)

type BookRepository struct {
    db *sql.DB
}

// Ensure BookRepository implements BookRepositoryInterface
var _ BookRepositoryInterface = (*BookRepository)(nil)

func NewBookRepository(cfg *config.Config) (*BookRepository, error) {
	connStr := cfg.GetDBConnectionString()
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // Create table with PostgreSQL syntax
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS books (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        price DECIMAL(10,2) NOT NULL,
        stock INTEGER NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
    CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
    `

    if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
        return nil, fmt.Errorf("failed to create table: %w", err)
    }

    return &BookRepository{db: db}, nil
}

func (r *BookRepository) Close() error {
    return r.db.Close()
}

func (r *BookRepository) Create(ctx context.Context, book *models.Book) error {
    // PostgreSQL uses RETURNING to get auto-generated ID and timestamps
    query := `INSERT INTO books (title, author, price, stock) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
    
    err := r.db.QueryRowContext(ctx, query, book.Title, book.Author, book.Price, book.Stock).Scan(
        &book.ID, &book.CreatedAt, &book.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create book: %w", err)
    }
    
    return nil
}

func (r *BookRepository) GetByID(ctx context.Context, id int) (*models.Book, error) {
    query := `SELECT id, title, author, price, stock, created_at, updated_at FROM books WHERE id = $1`
    var book models.Book
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &book.ID, &book.Title, &book.Author, &book.Price, &book.Stock,
        &book.CreatedAt, &book.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get book: %w", err)
    }
    return &book, nil
}

func (r *BookRepository) GetAll(ctx context.Context) ([]models.Book, error) {
    query := `SELECT id, title, author, price, stock, created_at, updated_at FROM books ORDER BY id DESC`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query books: %w", err)
    }
    defer func() {
		if err := rows.Close(); err != nil {
			fmt.Print("Failed to close rows")
		}
	}()
    var books []models.Book
    for rows.Next() {
        var book models.Book
        if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Price, &book.Stock,
            &book.CreatedAt, &book.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan book: %w", err)
        }
        books = append(books, book)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }

    return books, nil
}

func (r *BookRepository) Update(ctx context.Context, id int, book *models.Book) error {
    // PostgreSQL uses $1, $2 placeholders and RETURNING to confirm update
    query := `UPDATE books SET title = $1, author = $2, price = $3, stock = $4 WHERE id = $5 RETURNING updated_at`
    
    err := r.db.QueryRowContext(ctx, query, book.Title, book.Author, book.Price, book.Stock, id).Scan(
        &book.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return fmt.Errorf("book not found")
    }
    if err != nil {
        return fmt.Errorf("failed to update book: %w", err)
    }

    book.ID = id
    return nil
}

func (r *BookRepository) Delete(ctx context.Context, id int) error {
    query := `DELETE FROM books WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete book: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("book not found")
    }

    return nil
}

func (r *BookRepository) Health(ctx context.Context) error {
    return r.db.PingContext(ctx)
}

// GetStats returns database statistics (optional)
func (r *BookRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    
    // Get total book count
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&count)
    if err != nil {
        return nil, fmt.Errorf("failed to get book count: %w", err)
    }
    stats["total_books"] = count
    
    // Get total stock value
    var totalStock int
    err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(stock), 0) FROM books").Scan(&totalStock)
    if err == nil {
        stats["total_stock"] = totalStock
    }
    
    // Get average price
    var avgPrice float64
    err = r.db.QueryRowContext(ctx, "SELECT COALESCE(AVG(price), 0) FROM books").Scan(&avgPrice)
    if err == nil {
        stats["average_price"] = avgPrice
    }
    
    // Get database connection pool stats
    dbStats := r.db.Stats()
    stats["max_open_connections"] = dbStats.MaxOpenConnections
    stats["open_connections"] = dbStats.OpenConnections
    stats["in_use"] = dbStats.InUse
    stats["idle"] = dbStats.Idle
    stats["wait_count"] = dbStats.WaitCount
    stats["wait_duration"] = dbStats.WaitDuration.String()
    
    return stats, nil
}