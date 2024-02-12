package repository

import (
	"github.com/msmkdenis/wb-order-nats/internal/storage/db"
	"go.uber.org/zap"
)

type OrderRepository struct {
	postgresPool *db.PostgresPool
	logger       *zap.Logger
}

func NewOrderRepository(postgresPool *db.PostgresPool, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		postgresPool: postgresPool,
		logger:       logger,
	}
}
