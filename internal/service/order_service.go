package service

import (
	"context"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/model"
)

type OrderRepository interface {
	Insert(ctx context.Context, order model.Order) error
	SelectById(ctx context.Context, orderId string) (*model.Order, error)
	SelectAll(ctx context.Context) ([]model.Order, error)
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

func (o *OrderUseCase) Save(ctx context.Context, order model.Order) error {
	err := o.repository.Insert(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (o *OrderUseCase) FindById(ctx context.Context, orderId string) (*model.Order, error) {
	return o.repository.SelectById(ctx, orderId)
}

func (o *OrderUseCase) FindAll(ctx context.Context) ([]model.Order, error) {
	return o.repository.SelectAll(ctx)
}
