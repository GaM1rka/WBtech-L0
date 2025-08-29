package cache

import (
	"readermicroservice/internal/database"
	"readermicroservice/internal/models"
	"sync"
)

type Cache struct {
	mu       sync.RWMutex
	Elements map[string]models.Order
}

func New() *Cache {
	return &Cache{
		Elements: make(map[string]models.Order),
	}
}

func (c *Cache) Add(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Elements[order.OrderUID] = order
}

func (c *Cache) ResetDB(db *database.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	orders, err := db.GetAll()
	if err != nil {
		for _, item := range orders {
			c.Elements[item.OrderUID] = item
		}
	}
}
