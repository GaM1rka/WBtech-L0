package cache

import "readermicroservice/internal/models"

type OrderCache interface {
	Add(order models.Order)
	Get(orderUID string) (models.Order, bool)
	GetStats() (int, int, []string)
	StopCleanup()
	ResetDB(db OrderDatabase)
}

type OrderDatabase interface {
	GetAll() ([]models.Order, error)
}
