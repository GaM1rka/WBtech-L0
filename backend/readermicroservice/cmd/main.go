package main

import (
	"net/http"
	"readermicroservice/configs"
	"readermicroservice/internal"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/database"
	"readermicroservice/internal/kafka/consumer"
	"readermicroservice/internal/models"
	"time"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обрабатываем preflight OPTIONS запросы
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Middleware для логирования всех запросов
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs.RLogger.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		configs.RLogger.Printf("Headers: %v", r.Header)

		next(w, r)
	}
}

func main() {
	configs.Configure()

	cfg := models.Config{
		Host:     "db",
		Port:     "5432",
		User:     "myuser",
		Password: "mypassword",
		DBName:   "mydatabase",
	}
	time.Sleep(5 * time.Second)
	db, err := database.New(cfg)
	if err != nil {
		configs.RLogger.Println("Error while creating database: ", err)
	}
	err = db.CreateTables()

	if err != nil {
		configs.RLogger.Println("Error while creating tables in DB: ", err)
	}

	c := cache.New()
	c.ResetDB(db)

	go consumer.Listen(c, db)

	handler := internal.NewHandler(c, db)

	wrappedHandler := enableCORS(loggingMiddleware(handler.OrderHandler))

	http.HandleFunc("/order/", wrappedHandler)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		return
	}
}
