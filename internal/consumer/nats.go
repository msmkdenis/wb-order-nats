package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/msmkdenis/wb-order-nats/internal/metrics"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
	"time"
)

type OrderService interface {
	Save(ctx context.Context, order model.Order) error
}

type StatisticsPusher interface {
	PushStats(message metrics.MessageStat)
}

type NatsClient struct {
	client     stan.Conn
	os         OrderService
	sp         StatisticsPusher
	logger     *zap.Logger
	ordersChan chan model.Order
}

func NewNatsClient(cluster string, clientID string, natsURL string, service OrderService, sp StatisticsPusher, logger *zap.Logger) (*NatsClient, error) {
	client, err := stan.Connect(cluster, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Info("error", zap.Error(err))
		return nil, err
	}

	return &NatsClient{
		client:     client,
		os:         service,
		sp:         sp,
		logger:     logger,
		ordersChan: make(chan model.Order),
	}, nil
}

func (n *NatsClient) OrderProcessingRun() error {
	for i := 0; i < 5; i++ {
		go func() {
			ss, err := n.client.QueueSubscribe("orders", "test-queue", n.consumeOrder(), stan.DurableName("test-durable"),
				stan.DeliverAllAvailable(), stan.MaxInflight(20))
			fmt.Println(ss.IsValid())
			if err != nil {
				n.logger.Info("error", zap.Error(err))
				if n.client.Close() != nil {
					n.logger.Info("error", zap.Error(err))
				}
			}
		}()
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
					go func(order model.Order) {
						m := metrics.MessageStat{
							Id:        order.OrderUID,
							Status:    "error",
							Message:   err.Error(),
							Processed: time.Now().UTC(),
						}
						n.sp.PushStats(m)
						n.logger.Info("error", zap.Error(err))
					}(order)
				}
				n.logger.Info("saved", zap.String("id", order.OrderUID))
				go func(order model.Order) {
					m := metrics.MessageStat{
						Id:        order.OrderUID,
						Status:    "processed",
						Message:   "ok",
						Processed: time.Now().UTC(),
					}
					n.sp.PushStats(m)
				}(order)
			}
		}()
	}
}
