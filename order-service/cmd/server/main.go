package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/clients"
	"order-service/internal/config"
	"order-service/internal/handlers"
	"order-service/internal/middleware"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/pkg/logger"

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
	repo, err := repository.NewOrderRepository(cfg)
	if err != nil {
		log.Fatal("Failed to initialize repository", "error", err)
	}
	defer repo.Close()

	// Initialize clients
	bookClient := clients.NewBookClient(cfg.BookServiceURL, cfg.ServiceTimeout, log)
	userClient := clients.NewUserClient(cfg.UserServiceURL, cfg.ServiceTimeout, log)

	// Initialize service
	orderService := service.NewOrderService(repo, bookClient, userClient, log)

	// Initialize handlers
	orderHandlers := handlers.NewOrderHandlers(orderService, userClient, log)

	// Setup router
	router := mux.NewRouter()

	// Order routes (all require auth)
	router.HandleFunc("/orders", orderHandlers.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", orderHandlers.GetOrder).Methods("GET")
	router.HandleFunc("/orders/user/{userId}", orderHandlers.GetUserOrders).Methods("GET")

	// Health check
	router.HandleFunc("/health", orderHandlers.Health).Methods("GET")
	
	// Metrics
	router.Handle("/metrics", promhttp.Handler())

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
		log.Info("  POST   /api/v1/orders              - Create new order")
		log.Info("  GET    /api/v1/orders/{id}         - Get order by ID")
		log.Info("  GET    /api/v1/orders/user/{userId} - Get user orders")
		log.Info("  GET    /health                     - Health check")
		log.Info("  GET    /metrics                    - Prometheus metrics")
		
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
