package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
	"readermicroservice/internal/handler"
	"readermicroservice/internal/kafka/consumer"
)

const configPath = "configs/main.yml"

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config.RLogger.Printf("Received %s request for %s from %s",
			r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}

func main() {
	config.Configure()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		config.RLogger.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DB)
	if err != nil {
		config.RLogger.Fatalf("Error creating database connection: %v", err)
	}
	defer db.Close()

	// Initialize cache
	cache := cache.New(&cfg.Cache)
	defer cache.StopCleanup()

	// Load cache from database
	cache.ResetDB(db)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize and start Kafka consumer
	consumer, err := consumer.New(cfg, cache, db)
	if err != nil {
		config.RLogger.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()

	go func() {
		if err := consumer.Listen(ctx); err != nil && err != context.Canceled {
			config.RLogger.Printf("Kafka consumer error: %v", err)
		}
	}()

	// Setup HTTP handlers
	h := handler.New(cache, db)
	wrappedHandler := enableCORS(loggingMiddleware(h.OrderHandler))
	http.HandleFunc("/order/", wrappedHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown setup
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	serverErr := make(chan error, 1)
	go func() {
		config.RLogger.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal
	select {
	case sig := <-signalCh:
		config.RLogger.Printf("Received signal: %v. Starting graceful shutdown...", sig)
	case err := <-serverErr:
		config.RLogger.Printf("Server error: %v. Starting graceful shutdown...", err)
	}

	// Graceful shutdown
	cancel() // Stop Kafka consumer

	// Print cache stats before shutdown
	if count, maxSize, keys := cache.GetStats(); count > 0 {
		config.RLogger.Printf("Cache stats: %d/%d elements", count, maxSize)
		if len(keys) > 5 {
			config.RLogger.Printf("First 5 cache keys: %v", keys[:5])
		} else {
			config.RLogger.Printf("Cache keys: %v", keys)
		}
	}

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		config.RLogger.Printf("HTTP server shutdown error: %v", err)
	}

	config.RLogger.Println("Graceful shutdown completed")
}
