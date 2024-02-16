package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/stan.go"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/metrics"
	"github.com/msmkdenis/wb-order-nats/internal/model"
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

func (n *NatsClient) OrderProcessingRun(subject string, qGroup string, durable string, subscribers int, workers int, unsubscribe chan struct{}) error {
	for i := 0; i < subscribers; i++ {
		go func() {
			n.logger.Info("subscribing", zap.String("subject", subject), zap.String("qGroup", qGroup), zap.String("durable", durable))
			sc, err := n.client.QueueSubscribe(subject, qGroup, n.consumeOrder(), stan.DurableName(durable),
				stan.DeliverAllAvailable(), stan.MaxInflight(20))
			if err != nil {
				n.logger.Info("error", zap.Error(err))
				if n.client.Close() != nil {
					n.logger.Info("error", zap.Error(err))
				}
			}

			go func() {
				<-unsubscribe
				n.logger.Info("unsubscribing...")
				err = sc.Unsubscribe()
				if err != nil {
					n.logger.Info("error", zap.Error(err))
				}
			}()
		}()
	}
	go n.WorkerSaveRun(workers)
	return nil
}

func (n *NatsClient) consumeOrder() stan.MsgHandler {
	return func(msg *stan.Msg) {
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			n.logger.Info("error", zap.Error(err))
		} else {
			n.ordersChan <- order
		}
	}
}

func (n *NatsClient) WorkerSaveRun(workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for order := range n.ordersChan {
				err := n.os.Save(context.Background(), order)
				if err != nil {
					n.logger.Info("error", zap.Error(err))
					go func(order model.Order) {
						m := metrics.MessageStat{
							ID:        order.OrderUID,
							Status:    "error",
							Message:   err.Error(),
							Processed: time.Now().UTC(),
						}
						n.sp.PushStats(m)
						n.logger.Info("error", zap.Error(err))
					}(order)
				} else {
					n.logger.Info("saved", zap.String("id", order.OrderUID))
					go func(order model.Order) {
						m := metrics.MessageStat{
							ID:        order.OrderUID,
							Status:    "success",
							Message:   "ok",
							Processed: time.Now().UTC(),
						}
						n.sp.PushStats(m)
					}(order)
				}
			}
		}()
	}
}
