package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/metrics"
)

type StatisticsGetter interface {
	GetStats() map[string][]metrics.MessageStat
	GetStatsCount() *metrics.MessageStatCount
}

type StatCountsDTO struct {
	Processed int `json:"processed"`
	Failed    int `json:"failed"`
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
	e.GET("/api/v1/stats/counts", handler.GetStatsCount)

	return handler
}

func (h *StatisticsHandler) GetStats(c echo.Context) error {
	stat := h.statGetter.GetStats()

	return c.JSON(http.StatusOK, stat)
}

func (h *StatisticsHandler) GetStatsCount(c echo.Context) error {
	statCounts := h.statGetter.GetStatsCount()
	answer := StatCountsDTO{
		Processed: int(statCounts.Processed.Load()),
		Failed:    int(statCounts.Failed.Load()),
	}

	return c.JSON(http.StatusOK, answer)
}
