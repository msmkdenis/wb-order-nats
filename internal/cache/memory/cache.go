package memory

import (
	"sync"

	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/model"
)

type Cache struct {
	mu     sync.RWMutex
	items  map[string]model.Order
	logger *zap.Logger
}

func NewCache(logger *zap.Logger) *Cache {
	return &Cache{
		mu:     sync.RWMutex{},
		logger: logger,
	}
}

func (c *Cache) SetOrder(key string, value model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
}

func (c *Cache) GetOrder(key string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.items[key]

	c.logger.Info("Get from cache", zap.String("key", key))
	c.logger.Info("Value", zap.Any("value", value))

	return value, ok
}

func (c *Cache) RestoreCache(orders []model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	items := make(map[string]model.Order, len(orders))
	for _, order := range orders {
		items[order.OrderUID] = order
	}

	c.items = items
}
