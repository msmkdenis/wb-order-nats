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

type CacheSetter interface {
	SetOrder(key string, value model.Order)
	RestoreCache(orders []model.Order)
}

type OrderUseCase struct {
	repository OrderRepository
	cache      CacheSetter
	logger     *zap.Logger
}

func NewOrderUseCase(repository OrderRepository, cache CacheSetter, logger *zap.Logger) *OrderUseCase {
	return &OrderUseCase{
		repository: repository,
		cache:      cache,
		logger:     logger,
	}
}

func (o *OrderUseCase) Save(ctx context.Context, order model.Order) error {
	err := o.repository.Insert(ctx, order)
	if err != nil {
		return err
	}

	o.cache.SetOrder(order.OrderUID, order)
	return nil
}

func (o *OrderUseCase) FindById(ctx context.Context, orderId string) (*model.Order, error) {
	order, err := o.repository.SelectById(ctx, orderId)
	if err != nil {
		return nil, err
	}

	o.cache.SetOrder(orderId, *order)
	return order, nil
}

func (o *OrderUseCase) FindAll(ctx context.Context) ([]model.Order, error) {
	return o.repository.SelectAll(ctx)
}

func (o *OrderUseCase) RestoreCache() error {
	orders, err := o.repository.SelectAll(context.Background())
	if err != nil {
		return err
	}

	o.cache.RestoreCache(orders)
	return nil
}
