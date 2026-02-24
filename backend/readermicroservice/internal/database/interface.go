package database

import "readermicroservice/internal/models"

type Database interface {
	Insert(data models.Order) error
	GetByUID(orderUID string) (*models.Order, error)
	GetAll() ([]models.Order, error)
	Close() error
	Ping() error
}
