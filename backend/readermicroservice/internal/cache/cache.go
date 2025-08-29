package cache

import (
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

func (c *Cache) ResetDB() {
	c.mu.Lock()
	defer c.mu.Unlock()

}
