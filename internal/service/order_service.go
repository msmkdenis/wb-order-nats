package service

import (
	"go.uber.org/zap"
)

type OrderRepository interface {
}

type OrderUseCase struct {
	repository OrderRepository
	logger     *zap.Logger
}

func NewOrderUseCase(repository OrderRepository, logger *zap.Logger) *OrderUseCase {
	return &OrderUseCase{
		repository: repository,
		logger:     logger,
	}
}
