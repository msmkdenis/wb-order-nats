package handlers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/metrics"
)

type MessageStat struct {
	ID       string                `json:"id"`
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
	stats := make([]MessageStat, 0, len(stat))
	for key, value := range stat {
		stats = append(stats, MessageStat{
			ID:       key,
			Messages: value,
		})
	}

	return c.JSON(200, stats)
}
