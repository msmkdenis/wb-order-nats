package fakeproducer

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func Run(cluster string, clientID string, natsURL string) {
	logger, _ := zap.NewProduction()

	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			logger.Error("Error acking msg", zap.String("nuid", ackedNuid), zap.Error(err))
		} else {
			logger.Info("Msg acked", zap.String("nuid", ackedNuid))
		}
	}

	sc, err := stan.Connect(cluster, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Error("Error connecting", zap.Error(err))
		return
	}

	for i := 0; i < 3; i++ {
		order := newFakeOrder()
		or, _ := json.Marshal(order)
		_, err = sc.PublishAsync("orders", or, ackHandler) // returns immediately
		if err != nil {
			logger.Error("Error publishing", zap.Error(err))
		}
	}
}

func newFakeOrder() model.Order {
	faker := gofakeit.New(0)
	trackNumber := faker.Word()

	itemCount := faker.IntRange(1, 10)
	items := make([]model.Item, 0, itemCount)
	var totalSum int
	for i := 0; i < itemCount; i++ {
		item := newFakeItem(trackNumber)
		totalSum += item.TotalPrice
		items = append(items, item)
	}

	delivery := newFakeDelivery()

	payment := newFakePayment()
	payment.GoodsTotal = totalSum

	dateTime := faker.DateRange(time.Now().AddDate(0, -6, 0), time.Now())
	date := dateTime.Format("2006-01-02T15:04:05Z")

	return model.Order{
		OrderUID:          faker.UUID(),
		TrackNumber:       trackNumber,
		Entry:             trackNumber,
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            faker.LanguageAbbreviation(),
		InternalSignature: "",
		CustomerID:        faker.UUID(),
		DeliveryService:   faker.Word(),
		Shardkey:          strconv.Itoa(faker.IntRange(1, 9)),
		SmID:              faker.IntRange(1, 99),
		DateCreated:       date,
		OofShard:          strconv.Itoa(faker.IntRange(1, 9)),
	}
}

func newFakeItem(trackNumber string) model.Item {
	faker := gofakeit.New(0)
	price := faker.IntRange(1, 2000)
	sale := faker.IntRange(1, 30)
	totalPrice := price * (100 - sale) / 100
	return model.Item{
		ChrtID:      faker.IntRange(1111111, 9999999),
		TrackNumber: trackNumber,
		Price:       price,
		Rid:         faker.UUID(),
		Name:        faker.Word(),
		Sale:        sale,
		Size:        "0",
		TotalPrice:  totalPrice,
		NmID:        faker.IntRange(1111111, 9999999),
		Brand:       faker.Word(),
		Status:      202,
	}
}

func newFakePayment() model.Payment {
	faker := gofakeit.New(0)
	return model.Payment{
		Transaction:  faker.Word(),
		RequestID:    "",
		Currency:     faker.Currency().Short,
		Provider:     faker.Word(),
		Amount:       faker.IntRange(1, 2000),
		PaymentDt:    faker.DateRange(time.Now().AddDate(0, -6, 0), time.Now()).Unix(),
		Bank:         faker.RandomString([]string{"alpha", "sberbank", "sovcombank"}),
		DeliveryCost: faker.IntRange(1000, 3000),
		GoodsTotal:   0,
		CustomFee:    faker.IntRange(1000, 3000),
	}
}

func newFakeDelivery() model.Delivery {
	faker := gofakeit.New(0)
	return model.Delivery{
		Name:    faker.Name(),
		Phone:   fmt.Sprintf("+%s", faker.Phone()),
		Zip:     faker.Zip(),
		City:    faker.City(),
		Address: faker.Street(),
		Region:  faker.State(),
		Email:   faker.Email(),
	}
}
