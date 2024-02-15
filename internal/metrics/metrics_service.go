package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type MessageStat struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Processed time.Time `json:"processed"`
}

type MessageStatCount struct {
	Processed atomic.Int64 `json:"processed"`
	Failed    atomic.Int64 `json:"failed"`
}

type MessageStatsUseCase struct {
	statistics        map[string][]MessageStat
	mu                *sync.RWMutex
	logger            *zap.Logger
	processedMessages chan MessageStat
	msc               *MessageStatCount
}

func NewMessageStatsUseCase(logger *zap.Logger) *MessageStatsUseCase {
	return &MessageStatsUseCase{
		statistics:        make(map[string][]MessageStat),
		mu:                &sync.RWMutex{},
		logger:            logger,
		processedMessages: make(chan MessageStat),
		msc:               &MessageStatCount{},
	}
}

func (m *MessageStatsUseCase) PushStats(message MessageStat) {
	m.processedMessages <- message
}

func (m *MessageStatsUseCase) ProcessedMessagesRun(ctx context.Context) {
	for msg := range m.processedMessages {
		m.mu.Lock()
		m.statistics[msg.ID] = append(m.statistics[msg.ID], msg)
		m.mu.Unlock()
		m.msc.Processed.Add(1)
		if msg.Status == "error" {
			m.msc.Failed.Add(1)
		}
	}
}

func (m *MessageStatsUseCase) GetStats() map[string][]MessageStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.statistics

	return stats
}

func (m *MessageStatsUseCase) GetStatsCount() *MessageStatCount {
	return m.msc
}
