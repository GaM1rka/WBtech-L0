package cache

import (
	"sync"
	"time"

	"readermicroservice/internal/config"
	"readermicroservice/internal/models"
)

type cacheItem struct {
	order      models.Order
	lastAccess time.Time
	createdAt  time.Time
}

// Cache реализует интерфейс OrderCache
type Cache struct {
	mu              sync.RWMutex
	elements        map[string]*cacheItem
	maxSize         int
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// New создает новый экземпляр кэша
func New(conf *config.CacheConfig) *Cache {
	c := &Cache{
		elements:        make(map[string]*cacheItem),
		maxSize:         conf.MaxSize,
		defaultTTL:      conf.DefaultTTL,
		cleanupInterval: conf.CleanupInterval,
		stopCleanup:     make(chan bool),
	}

	go c.startCleanup()
	return c
}

// Add добавляет элемент в кэш
func (c *Cache) Add(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.elements) >= c.maxSize {
		c.evictOldest()
	}

	c.elements[order.OrderUID] = &cacheItem{
		order:      order,
		lastAccess: time.Now(),
		createdAt:  time.Now(),
	}
}

// Get получает элемент из кэша
func (c *Cache) Get(orderUID string) (models.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.elements[orderUID]
	if !exists {
		return models.Order{}, false
	}

	item.lastAccess = time.Now()
	return item.order, true
}

// evictOldest удаляет самый старый по доступу элемент
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.elements {
		if oldestKey == "" || item.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.lastAccess
		}
	}

	if oldestKey != "" {
		delete(c.elements, oldestKey)
	}
}

// startCleanup запускает периодическую очистку просроченных элементов
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpired удаляет просроченные элементы
func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.elements {
		if now.Sub(item.createdAt) > c.defaultTTL {
			delete(c.elements, key)
		}
	}
}

// StopCleanup останавливает горутину очистки
func (c *Cache) StopCleanup() {
	close(c.stopCleanup)
}

// GetStats возвращает статистику кэша
func (c *Cache) GetStats() (int, int, []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.elements))
	for k := range c.elements {
		keys = append(keys, k)
	}
	return len(c.elements), c.maxSize, keys
}

// ResetDB загружает данные из БД в кэш
func (c *Cache) ResetDB(db OrderDatabase) {
	c.mu.Lock()
	defer c.mu.Unlock()

	orders, err := db.GetAll()
	if err != nil {
		config.RLogger.Printf("Error loading data from DB to cache: %v", err)
		return
	}

	c.elements = make(map[string]*cacheItem)

	for _, item := range orders {
		if len(c.elements) < c.maxSize {
			c.elements[item.OrderUID] = &cacheItem{
				order:      item,
				lastAccess: time.Now(),
				createdAt:  time.Now(),
			}
		}
	}
}
