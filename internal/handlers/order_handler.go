package handlers

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"go.uber.org/zap"
	"net/http"
)

type OrderService interface {
	Save(ctx context.Context, order model.Order) error
	FindById(ctx context.Context, orderId string) (*model.Order, error)
	FindAll(ctx context.Context) ([]model.Order, error)
}

type OrderHandler struct {
	orderService OrderService
	logger       *zap.Logger
}

func NewOrderHandler(e *echo.Echo, service OrderService, logger *zap.Logger) *OrderHandler {
	handler := &OrderHandler{
		orderService: service,
		logger:       logger,
	}

	e.POST("/api/v1/order", handler.SaveOrder)
	e.GET("/api/v1/order/:orderId", handler.FindOrderById)
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

func (h *OrderHandler) FindOrderById(c echo.Context) error {
	orderId := c.Param("orderId")

	order, err := h.orderService.FindById(context.Background(), orderId)
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
