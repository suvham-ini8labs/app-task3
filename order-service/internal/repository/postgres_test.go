package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"order-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	order := &models.Order{
		UserID:     1,
		BookID:     101,
		Quantity:   2,
		TotalPrice: 50.0,
		Status:     models.StatusConfirmed,
	}

	expectedID := 1
	expectedCreatedAt := time.Now()
	expectedUpdatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(expectedID, expectedCreatedAt, expectedUpdatedAt)

	mock.ExpectQuery("INSERT INTO orders").
		WithArgs(order.UserID, order.BookID, order.Quantity, order.TotalPrice, order.Status).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), order)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, order.ID)
	assert.Equal(t, expectedCreatedAt, order.CreatedAt)
	assert.Equal(t, expectedUpdatedAt, order.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	expectedID := 1
	rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at", "updated_at"}).
		AddRow(expectedID, 1, 101, 2, 50.0, models.StatusConfirmed, time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE id = \\$1").
		WithArgs(expectedID).
		WillReturnRows(rows)

	order, err := repo.GetByID(context.Background(), expectedID)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, expectedID, order.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE id = \\$1").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	order, err := repo.GetByID(context.Background(), 999)

	assert.NoError(t, err)
	assert.Nil(t, order)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	userID := 1
	rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at", "updated_at"}).
		AddRow(1, userID, 101, 1, 20.0, models.StatusConfirmed, time.Now(), time.Now()).
		AddRow(2, userID, 102, 1, 30.0, models.StatusConfirmed, time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	orders, err := repo.GetByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	orderID := 1
	status := models.StatusCancelled

	mock.ExpectQuery("UPDATE orders SET status = \\$1, updated_at = CURRENT_TIMESTAMP WHERE id = \\$2").
		WithArgs(status, orderID).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(time.Now()))

	err = repo.UpdateStatus(context.Background(), orderID, status)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	orderID := 1

	mock.ExpectExec("DELETE FROM orders WHERE id = \\$1").
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), orderID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Health_Healthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	mock.ExpectPing()

	err = repo.Health(context.Background())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Health_Unhealthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := &OrderRepository{db: db}

	mock.ExpectPing().WillReturnError(errors.New("db down"))

	err = repo.Health(context.Background())

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
