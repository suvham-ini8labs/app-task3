package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"order-service/internal/clients"
	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gorilla/mux"
)

type OrderHandlers struct {
	service    service.OrderServiceInterface
	userClient clients.UserServiceClient
	logger     Logger
}

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

func NewOrderHandlers(service service.OrderServiceInterface, userClient clients.UserServiceClient, logger Logger) *OrderHandlers {
	return &OrderHandlers{
		service:    service,
		userClient: userClient,
		logger:     logger,
	}
}

func (h *OrderHandlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Extract token and user ID
	token, userID, err := h.extractTokenAndUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.BookID <= 0 {
		sendError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}
	if req.Quantity <= 0 {
		sendError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	order, err := h.service.CreateOrder(ctx, userID, token, &req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, order)
}

func (h *OrderHandlers) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Extract token and user ID
	token, userID, err := h.extractTokenAndUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Get order ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := h.service.GetOrder(ctx, id, token)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "order not found" {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	// Verify order belongs to user
	if order.UserID != userID {
		sendError(w, http.StatusForbidden, "Access denied: order does not belong to user")
		return
	}

	sendJSON(w, http.StatusOK, order)
}

func (h *OrderHandlers) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Extract token and user ID
	token, authUserID, err := h.extractTokenAndUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	fmt.Println(authUserID, userID)
	// Verify user is accessing their own orders
	if authUserID != userID {
		sendError(w, http.StatusForbidden, "Access denied: can only view your own orders")
		return
	}

	orders, err := h.service.GetUserOrders(ctx, userID, token)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, orders)
}

func (h *OrderHandlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.Health(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *OrderHandlers) extractTokenAndUserID(r *http.Request) (string, int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", 0, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", 0, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]
	
	// Validate token with user service and get user info
	user, err := h.userClient.ValidateToken(r.Context(), token)
	if err != nil {
		return "", 0, fmt.Errorf("invalid or expired token: %w", err)
	}

	return token, user.ID, nil
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, models.ErrorResponse{
		Error: message,
		Code:  status,
	})
}
