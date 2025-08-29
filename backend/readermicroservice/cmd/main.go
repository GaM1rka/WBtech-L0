package main

import (
	"net/http"
	"readermicroservice/configs"
	"readermicroservice/internal"
	"readermicroservice/internal/cache"
	"readermicroservice/internal/database"
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
		User:     "testUser",
		Password: "testPassword",
		DBName:   "testDBName",
	}
	db, err := database.New(cfg)
	if err != nil {
		configs.RLogger.Println("Error while creating database: ", err)
	}
	c := cache.New()
	c.ResetDB(db)

	handler := internal.NewHandler(c, db)

	http.HandleFunc("/order/", handler.OrderHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return
	}
}
