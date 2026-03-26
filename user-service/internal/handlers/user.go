package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"user-service/internal/middleware"
	"user-service/internal/models"
	"user-service/internal/service"
	"user-service/pkg/logger"

	"github.com/gorilla/mux"
)

type UserHandlers struct {
	service service.UserService
	logger  *logger.Logger
}

func NewUserHandlers(service service.UserService, log *logger.Logger) *UserHandlers {
	return &UserHandlers{
		service: service,
		logger:  log,
	}
}

func (h *UserHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", "error", err)
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "email already registered" {
			status = http.StatusConflict
		}
		sendError(w, status, err.Error())
		return
	}

	h.logger.Info("User registered via API", "id", user.ID, "email", user.Email)
	sendJSON(w, http.StatusCreated, user)
}

func (h *UserHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode login request", "error", err)
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Login(r.Context(), &req)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.logger.Info("User logged in via API", "email", req.Email)
	sendJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Error("Invalid user ID in URL", "error", err, "id", vars["id"])
		sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user ID from context (from auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("Unauthorized access attempt", "user_id", id)
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Verify user is accessing their own profile
	if userID != id {
		h.logger.Warn("Access denied: user tried to access another user's profile", 
			"requested_id", id, "authenticated_id", userID)
		sendError(w, http.StatusForbidden, "Access denied: you can only access your own profile")
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, user)
}

func (h *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Error("Invalid user ID in URL", "error", err, "id", vars["id"])
		sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user ID from context (from auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error("Unauthorized update attempt", "user_id", id)
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Verify user is updating their own profile
	if userID != id {
		h.logger.Warn("Access denied: user tried to update another user's profile",
			"requested_id", id, "authenticated_id", userID)
		sendError(w, http.StatusForbidden, "Access denied: you can only update your own profile")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update request", "error", err, "user_id", id)
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), id, &req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	h.logger.Info("User updated via API", "id", id)
	sendJSON(w, http.StatusOK, user)
}

func (h *UserHandlers) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Health(r.Context()); err != nil {
		h.logger.Error("Health check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}
