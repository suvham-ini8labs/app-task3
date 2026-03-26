package repository

import (
    "context"
    "database/sql"
    "time"
    "fmt"
    "order-service/internal/models"
    "order-service/internal/config"

    _ "github.com/lib/pq"
)

type OrderRepository struct {
    db *sql.DB
}

var _ OrderRepositoryInterface = (*OrderRepository)(nil)

func NewOrderRepository(cfg *config.Config) (*OrderRepository, error) {
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
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL,
        book_id INTEGER NOT NULL,
        quantity INTEGER NOT NULL,
        total_price DECIMAL(10,2) NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
    CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
    CREATE INDEX IF NOT EXISTS idx_orders_book_id ON orders(book_id);
    `

    if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
        return nil, fmt.Errorf("failed to create table: %w", err)
    }

    return &OrderRepository{db: db}, nil
}

func (r *OrderRepository) Close() error {
    return r.db.Close()
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
    // PostgreSQL uses RETURNING to get auto-generated ID and timestamps
    query := `INSERT INTO orders (user_id, book_id, quantity, total_price, status) 
              VALUES ($1, $2, $3, $4, $5) 
              RETURNING id, created_at, updated_at`
    
    err := r.db.QueryRowContext(ctx, query, order.UserID, order.BookID, 
        order.Quantity, order.TotalPrice, order.Status).Scan(
        &order.ID, &order.CreatedAt, &order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    
    return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int) (*models.Order, error) {
    query := `SELECT id, user_id, book_id, quantity, total_price, status, created_at, updated_at 
              FROM orders WHERE id = $1`
    
    var order models.Order
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &order.ID, &order.UserID, &order.BookID, &order.Quantity, 
        &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    
    return &order, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int) ([]models.Order, error) {
    query := `SELECT id, user_id, book_id, quantity, total_price, status, created_at, updated_at 
              FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
    
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query orders: %w", err)
    }
   
    defer func() {
		if err := rows.Close(); err != nil {
			fmt.Print("Failed to close rows", "error", err)
		}
	}()

    var orders []models.Order
    for rows.Next() {
        var order models.Order
        if err := rows.Scan(
            &order.ID, &order.UserID, &order.BookID, &order.Quantity,
            &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        orders = append(orders, order)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }
    
    return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int, status models.OrderStatus) error {
    // PostgreSQL uses $1, $2 placeholders and RETURNING to confirm update
    query := `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING updated_at`
    
    var updatedAt time.Time
    err := r.db.QueryRowContext(ctx, query, status, id).Scan(&updatedAt)
    
    if err == sql.ErrNoRows {
        return fmt.Errorf("order not found")
    }
    if err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }
    
    return nil
}

func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
    query := `UPDATE orders SET user_id = $1, book_id = $2, quantity = $3, total_price = $4, 
              status = $5 WHERE id = $6 RETURNING updated_at`
    
    err := r.db.QueryRowContext(ctx, query, order.UserID, order.BookID, order.Quantity,
        order.TotalPrice, order.Status, order.ID).Scan(&order.UpdatedAt)
    
    if err == sql.ErrNoRows {
        return fmt.Errorf("order not found")
    }
    if err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    return nil
}

func (r *OrderRepository) Delete(ctx context.Context, id int) error {
    query := `DELETE FROM orders WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete order: %w", err)
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("order not found")
    }
    
    return nil
}

func (r *OrderRepository) Health(ctx context.Context) error {
    return r.db.PingContext(ctx)
}

// GetStats returns database statistics (optional)
func (r *OrderRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    
    // Get total order count
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&count)
    if err != nil {
        return nil, fmt.Errorf("failed to get order count: %w", err)
    }
    stats["total_orders"] = count
    
    // Get orders by status
    rows, err := r.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM orders GROUP BY status")
    if err == nil {
        defer func() {
		    if err := rows.Close(); err != nil {
			    fmt.Print("Failed to close rows", "error", err)
		    }
	    }()
        
        statusCounts := make(map[string]int)
        for rows.Next() {
            var status string
            var cnt int
            if err := rows.Scan(&status, &cnt); err == nil {
                statusCounts[status] = cnt
            }
        }
        stats["orders_by_status"] = statusCounts
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