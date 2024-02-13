package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

type OrderService interface {
	Save(ctx context.Context, order model.Order) error
}

type NatsClient struct {
	client stan.Conn
	os     OrderService
	logger *zap.Logger
}

func NewNatsClient(cluster string, clientID string, natsURL string, service OrderService, logger *zap.Logger) (*NatsClient, error) {
	client, err := stan.Connect(cluster, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Info("error", zap.Error(err))
		return nil, err
	}

	return &NatsClient{
		client: client,
		os:     service,
		logger: logger,
	}, nil
}

func (n *NatsClient) OrderProcessingRun() error {
	_, err := n.client.Subscribe("orders", n.consumeOrder(), stan.DurableName("test-durable"), stan.MaxInflight(20))
	if err != nil {
		n.logger.Info("error", zap.Error(err))
		return err
	}
	return nil
}

func (n *NatsClient) consumeOrder() stan.MsgHandler {
	return func(msg *stan.Msg) {
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			n.logger.Info("error", zap.Error(err))
		} else {
			item, _ := json.MarshalIndent(order, "", "  ")
			fmt.Println(string(item))
			err = n.os.Save(context.Background(), order)
			if err != nil {
				n.logger.Info("error", zap.Error(err))
			}
		}
	}
}
