package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"book-service/internal/config"
	"book-service/internal/handlers"
	"book-service/internal/middleware"
	"book-service/internal/repository"
	"book-service/internal/service"
	"book-service/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.LogLevel, "json")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer log.Sync()

	// Initialize repository
	repo, err := repository.NewBookRepository(cfg)
	if err != nil {
		log.Fatal("Failed to initialize repository", "error", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Error("Failed to close repository", "error", err)
		}
	}()
	
	// Initialize service
	bookService := service.NewBookService(repo)

	// Initialize handlers
	bookHandlers := handlers.NewBookHandlers(bookService)

	// Setup router
	router := mux.NewRouter()

	router.HandleFunc("/books/health", bookHandlers.Health).Methods("GET")
	router.Handle("/books/metrics", promhttp.Handler())
	
	// Book routes
	router.HandleFunc("/books", bookHandlers.CreateBook).Methods("POST")
	router.HandleFunc("/books", bookHandlers.ListBooks).Methods("GET")
	router.HandleFunc("/books/{id}", bookHandlers.GetBook).Methods("GET")
	router.HandleFunc("/books/{id}", bookHandlers.UpdateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", bookHandlers.DeleteBook).Methods("DELETE")

	// Health check
	
	// Metrics

	// Apply metrics middleware
	handler := middleware.Metrics(router)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Info("Server starting", "port", cfg.ServerPort)
		log.Info("Available endpoints:")
		log.Info("  POST   /api/v1/books          - Create book")
		log.Info("  GET    /api/v1/books          - List all books")
		log.Info("  GET    /api/v1/books/{id}     - Get book by ID")
		log.Info("  PUT    /api/v1/books/{id}     - Update book")
		log.Info("  DELETE /api/v1/books/{id}     - Delete book")
		log.Info("  GET    /health                - Health check")
		log.Info("  GET    /metrics               - Prometheus metrics")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}
