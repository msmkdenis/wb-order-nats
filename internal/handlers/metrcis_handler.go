package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/msmkdenis/wb-order-nats/internal/metrics"
	"go.uber.org/zap"
)

type MessageStat struct {
	Id       string                `json:"id"`
	Messages []metrics.MessageStat `json:"messages"`
}

type StatisticsGetter interface {
	GetStats() map[string][]metrics.MessageStat
}

type StatisticsHandler struct {
	statGetter StatisticsGetter
	logger     *zap.Logger
}

func NewStatisticsHandler(e *echo.Echo, service StatisticsGetter, logger *zap.Logger) *StatisticsHandler {
	handler := &StatisticsHandler{
		statGetter: service,
		logger:     logger,
	}

	e.GET("/api/v1/stats", handler.GetStats)

	return handler
}

func (h *StatisticsHandler) GetStats(c echo.Context) error {
	stat := h.statGetter.GetStats()
	var stats []MessageStat
	for key, value := range stat {
		stats = append(stats, MessageStat{
			Id:       key,
			Messages: value,
		})
	}

	return c.JSON(200, stats)
}
