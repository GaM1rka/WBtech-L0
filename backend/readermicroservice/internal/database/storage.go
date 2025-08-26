package database

import (
	"database/sql"
	"fmt"
	"readermicroservice/configs"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Config struct {
	Port     string
	User     string
	Password string
	DBName   string
}

func new(cfg Config) (*DB, error) {
	connStr := fmt.Sprintf("port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		configs.RLogger.Println("Error while initializing new DB: ", err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		configs.RLogger.Println("Falied to ping DB: ", err)
		return nil, err
	}
	return &DB{db}, nil
}
