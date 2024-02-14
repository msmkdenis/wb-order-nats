package middleware

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/model"
)

type CacheGetter interface {
	GetOrder(key string) (model.Order, bool)
}

type CacheMiddleware struct {
	cache  CacheGetter
	logger *zap.Logger
}

func NewCacheMiddleware(cache CacheGetter, logger *zap.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		cache:  cache,
		logger: logger,
	}
}

func (m *CacheMiddleware) GetFromCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			orderID := c.Param("orderID")
			order, ok := m.cache.GetOrder(orderID)
			if !ok {
				c.Response().Header().Set("X-Cache", "None")
				return next(c)
			}
			c.Response().Header().Set("Content-Type", "application/json")
			c.Response().Header().Set("X-Cache", "Cached")
			return c.JSON(200, order)
		}
	}
}
