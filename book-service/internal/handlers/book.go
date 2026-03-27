package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"fmt"
	"book-service/internal/models"
	"book-service/internal/service"

	"github.com/gorilla/mux"
)

type BookHandlers struct {
	service service.BookServiceInterface
}

func NewBookHandlers(service service.BookServiceInterface) *BookHandlers {
	return &BookHandlers{
		service: service,
	}
}

func (h *BookHandlers) CreateBook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req models.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	book, err := h.service.CreateBook(ctx, &req)
	if err != nil {
		status := http.StatusBadRequest
		sendError(w, status, err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, book)
}

func (h *BookHandlers) GetBook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := h.service.GetBook(ctx, id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "invalid book id" {
			status = http.StatusBadRequest
		}
		sendError(w, status, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, book)
}

func (h *BookHandlers) ListBooks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	books, err := h.service.ListBooks(ctx)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, books)
}

func (h *BookHandlers) UpdateBook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	var req models.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	book, err := h.service.UpdateBook(ctx, id, &req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "book not found" {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, book)
}

func (h *BookHandlers) DeleteBook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	if err := h.service.DeleteBook(ctx, id); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "book not found" {
			status = http.StatusNotFound
		}
		sendError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BookHandlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.Health(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()}); encodeErr != nil {
			fmt.Print("Failed to encode health response", "error", encodeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		fmt.Print("Failed to encode health response", "error", err)
	}
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
