package models

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusFailed    OrderStatus = "failed"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	BookID     int         `json:"book_id"`
	Quantity   int         `json:"quantity"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type CreateOrderRequest struct {
	BookID   int `json:"book_id"`
	Quantity int `json:"quantity"`
}

type OrderResponse struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	BookID     int         `json:"book_id"`
	Quantity   int         `json:"quantity"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	Book       *BookInfo   `json:"book,omitempty"`
	User       *UserInfo   `json:"user,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

type BookInfo struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
	Stock  int     `json:"stock"`
}

type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}
