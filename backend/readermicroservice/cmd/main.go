package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"readermicroservice/internal"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/config"
	"readermicroservice/internal/database"
	"readermicroservice/internal/kafka/consumer"
	"readermicroservice/internal/models"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

const (
	path string = "readermicroservice/configs/main.yml"
)

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
		config.RLogger.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		config.RLogger.Printf("Headers: %v", r.Header)

		next(w, r)
	}
}

func main() {
	config.Configure()

	// ПРОБЛЕМА 1: Не проверяем ошибку загрузки конфига
	var DBConfig *config.DBConfig
	DBConfig, err := config.LoadDBConfig(path)
	if err != nil {
		config.RLogger.Fatalf("Error loading DB config: %v", err)
	}

	cfg := models.Config{
		Host:     DBConfig.Host,
		Port:     DBConfig.Port,
		User:     DBConfig.User,
		Password: DBConfig.Password,
		Database: DBConfig.Database,
	}

	time.Sleep(5 * time.Second)

	// ПРОБЛЕМА 2: Если создание БД неуспешно, продолжаем выполнение
	db, err := database.New(cfg)
	if err != nil {
		config.RLogger.Fatalf("Error while creating database: %v", err) // Используем Fatalf вместо Println
	}

	err = db.CreateTables()
	if err != nil {
		config.RLogger.Fatalf("Error while creating tables in DB: %v", err)
	}

	c := cache.New()
	c.ResetDB(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Listen(ctx, c, db); err != nil && err != context.Canceled {
			config.RLogger.Printf("Kafka consumer error: %v", err)
		}
	}()

	handler := internal.NewHandler(c, db)
	wrappedHandler := enableCORS(loggingMiddleware(handler.OrderHandler))

	http.HandleFunc("/order/", wrappedHandler)

	server := &http.Server{
		Addr:         ":8081",
		Handler:      nil,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// ПРОБЛЕМА 3: Нет канала для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		config.RLogger.Printf("Starting server on %s", server.Addr) // Исправлено: "on" вместо "in"
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// ПРОБЛЕМА 4: Ожидаем только сигналы, но не ошибки сервера
	select {
	case sig := <-signalCh:
		config.RLogger.Printf("Received signal: %v. Starting graceful shutdown...", sig)
	case err := <-serverErr:
		config.RLogger.Printf("Server error: %v. Starting graceful shutdown...", err)
	}

	// ПРОБЛЕМА 5: Неправильная последовательность shutdown
	config.RLogger.Println("Starting graceful shutdown...")

	// Сначала отменяем контекст для остановки Kafka consumer
	cancel()

	// Даем немного времени для завершения обработки сообщений
	time.Sleep(2 * time.Second)

	// Затем останавливаем очистку кэша
	c.StopCleanup()

	// Логируем статистику кэша
	if count, maxSize, keys := c.GetStats(); count > 0 {
		config.RLogger.Printf("Cache stats: %d/%d elements", count, maxSize)
		// Логируем только первые несколько ключей для читаемости
		if len(keys) > 5 {
			config.RLogger.Printf("First 5 cache keys: %v", keys[:5])
		} else {
			config.RLogger.Printf("Cache keys: %v", keys)
		}
	}

	// Затем останавливаем HTTP сервер
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		config.RLogger.Printf("HTTP server shutdown error: %v", err)
	}

	config.RLogger.Println("Graceful shutdown completed")
}
