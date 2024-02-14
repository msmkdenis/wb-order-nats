package metrics

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
)

type MessageStat struct {
	Id        string    `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Processed time.Time `json:"processed"`
}

type MessageStatsUseCase struct {
	statistics        map[string][]MessageStat
	mu                sync.RWMutex
	logger            *zap.Logger
	processedMessages chan MessageStat
}

func NewMessageStatsUseCase(logger *zap.Logger) *MessageStatsUseCase {
	return &MessageStatsUseCase{
		statistics:        make(map[string][]MessageStat),
		mu:                sync.RWMutex{},
		logger:            logger,
		processedMessages: make(chan MessageStat),
	}
}

func (m *MessageStatsUseCase) PushStats(message MessageStat) {
	m.processedMessages <- message
}

func (m *MessageStatsUseCase) ProcessedMessagesRun(ctx context.Context) {

	for msg := range m.processedMessages {
		m.mu.Lock()
		if _, ok := m.statistics[msg.Id]; ok {
			m.statistics[msg.Id] = append(m.statistics[msg.Id], msg)
		} else {
			m.statistics[msg.Id] = []MessageStat{msg}
		}
		m.mu.Unlock()
	}

	select {
	case <-ctx.Done():
		m.logger.Info("collecting metrics shutdown", zap.Error(ctx.Err()))
		return
	}
}

func (m *MessageStatsUseCase) GetStats() map[string][]MessageStat {
	return m.statistics
}
