package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/internal/config"
	"user-service/internal/handlers"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/service"
	"user-service/pkg/logger"

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
	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer log.Sync()

	// Initialize repository
	repo, err := repository.NewPostgresUserRepository(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize repository", "error", err)
	}
	defer repo.Close()

	// Initialize service
	userService := service.NewUserService(repo, cfg.JWTSecret, cfg.JWTExpiration, log)

	// Initialize handlers
	userHandlers := handlers.NewUserHandlers(userService, log)

	// Setup router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/users", userHandlers.Register).Methods("POST")
	router.HandleFunc("/users/login", userHandlers.Login).Methods("POST")
	router.HandleFunc("/users/{id}", userHandlers.GetUser).Methods("GET")
	router.HandleFunc("/health", userHandlers.Health).Methods("GET")
	router.Handle("/metrics", promhttp.Handler())

	// Protected routes with auth middleware
	authMiddleware := middleware.Auth(userService, log)
	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware)
	protectedRouter.HandleFunc("/users/{id}", userHandlers.UpdateUser).Methods("PUT")

	// Apply metrics middleware
	handler := middleware.Metrics(router)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("User service starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down user service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("User service exited")
}
