package cache

import (
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"
	"sync"
)

type Cache struct {
	mu       sync.RWMutex
	elements map[string]models.Order
}

func New() *Cache {
	return &Cache{
		elements: make(map[string]models.Order),
	}
}

func (c *Cache) Add(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.elements[order.OrderUID] = order
}

func (c *Cache) ResetDB(db *database.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	orders, err := db.GetAll()
	if err != nil {
		for _, item := range orders {
			c.elements[item.OrderUID] = item
		}
	}
}
