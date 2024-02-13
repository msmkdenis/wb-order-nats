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
	client     stan.Conn
	os         OrderService
	logger     *zap.Logger
	ordersChan chan model.Order
}

func NewNatsClient(cluster string, clientID string, natsURL string, service OrderService, logger *zap.Logger) (*NatsClient, error) {
	client, err := stan.Connect(cluster, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Info("error", zap.Error(err))
		return nil, err
	}

	return &NatsClient{
		client:     client,
		os:         service,
		logger:     logger,
		ordersChan: make(chan model.Order),
	}, nil
}

func (n *NatsClient) OrderProcessingRun() error {
	ss, err := n.client.Subscribe("orders", n.consumeOrder(), stan.DurableName("test-durable"),
		stan.DeliverAllAvailable(), stan.MaxInflight(20))
	fmt.Println(ss.IsValid())
	if err != nil {
		n.logger.Info("error", zap.Error(err))
		return err
	}

	go n.WorkerSaveRun()
	return nil
}

func (n *NatsClient) consumeOrder() stan.MsgHandler {
	return func(msg *stan.Msg) {
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		fmt.Println("received not saved", order)
		if err != nil {
			n.logger.Info("error", zap.Error(err))
		} else {
			n.ordersChan <- order
		}
	}
}

func (n *NatsClient) WorkerSaveRun() {
	for i := 0; i < 100; i++ {
		go func() {
			for order := range n.ordersChan {
				err := n.os.Save(context.Background(), order)
				if err != nil {
					n.logger.Info("error", zap.Error(err))
				}
			}
		}()
	}
}
