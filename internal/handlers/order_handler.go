package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type OrderService interface {
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

	return handler
}
