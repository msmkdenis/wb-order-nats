package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"go.uber.org/zap"
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
			orderId := c.Param("orderId")
			order, ok := m.cache.GetOrder(orderId)
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
