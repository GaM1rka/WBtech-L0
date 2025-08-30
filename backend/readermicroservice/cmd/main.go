package main

import (
	"net/http"
	"readermicroservice/configs"
	"readermicroservice/internal"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/database"
	"readermicroservice/internal/kafka/consumer"
	"readermicroservice/internal/models"
)

type Config struct {
	Port     string
	User     string
	Password string
	DBName   string
}

func main() {
	configs.Configure()

	cfg := models.Config{
		Port:     "5432",
		User:     "myuser",
		Password: "mypassword",
		DBName:   "mydatabase",
	}
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

	http.HandleFunc("/order/", handler.OrderHandler)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		return
	}
}
