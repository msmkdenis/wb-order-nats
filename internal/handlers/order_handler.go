package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/middleware"
	"github.com/msmkdenis/wb-order-nats/internal/model"
)

type OrderService interface {
	Save(ctx context.Context, order model.Order) error
	FindByID(ctx context.Context, orderID string) (*model.Order, error)
	FindAll(ctx context.Context) ([]model.Order, error)
}

type OrderHandler struct {
	orderService OrderService
	cache        *middleware.CacheMiddleware
	logger       *zap.Logger
}

func NewOrderHandler(e *echo.Echo, service OrderService, cache *middleware.CacheMiddleware, logger *zap.Logger) *OrderHandler {
	handler := &OrderHandler{
		orderService: service,
		cache:        cache,
		logger:       logger,
	}

	e.POST("/api/v1/order", handler.SaveOrder)
	e.GET("/api/v1/order/:orderId", handler.FindOrderByID, cache.GetFromCache())
	e.GET("/api/v1/order/", handler.FindAll)

	return handler
}

func (h *OrderHandler) SaveOrder(c echo.Context) error {
	header := c.Request().Header.Get("Content-Type")
	if header != "application/json" {
		msg := "Content-Type header is not application/json"
		h.logger.Info("UnsupportedMediaType: " + msg)
		return c.JSON(http.StatusUnsupportedMediaType, map[string]string{"UnsupportedMediaType": "Content-Type header is not application/json"})
	}

	var order model.Order
	err := c.Bind(&order)
	if err != nil {
		h.logger.Info("Error while binding request", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error while binding request": err.Error()})
	}

	err = h.orderService.Save(context.TODO(), order)
	if err != nil {
		h.logger.Error("error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return c.JSON(200, order)
}

func (h *OrderHandler) FindOrderByID(c echo.Context) error {
	orderID := c.Param("orderID")

	order, err := h.orderService.FindByID(context.Background(), orderID)
	if err != nil {
		h.logger.Error("error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return c.JSON(200, order)
}

func (h *OrderHandler) FindAll(c echo.Context) error {
	orders, err := h.orderService.FindAll(context.Background())
	if err != nil {
		h.logger.Error("error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return c.JSON(200, orders)
}
