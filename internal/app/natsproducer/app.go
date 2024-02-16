package natsproducer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/model"
)

type producerConfig struct {
	Cluster    string
	Client     string
	NatsURL    string
	ServerAddr string
}

func Run() {
	err := godotenv.Load("natsproducer.env")
	if err != nil {
		log.Info("Error loading .env file, using default values")
	}

	config := &producerConfig{
		Cluster:    os.Getenv("NATS_PRODUCER_CLUSTER"),
		Client:     os.Getenv("NATS_PRODUCER_CLIENT"),
		NatsURL:    os.Getenv("NATS_PRODUCER_URL"),
		ServerAddr: os.Getenv("PRODUCER_SERV_ADDR"),
	}

	logger, _ := zap.NewProduction()
	producer := New(config.Cluster, config.Client, config.NatsURL, logger)
	e := echo.New()
	NewProducerHandler(e, producer, logger)

	errStart := e.Start(config.ServerAddr)
	if errStart != nil && !errors.Is(errStart, http.ErrServerClosed) {
		logger.Fatal(errStart.Error())
	}
}

type Producer struct {
	sc     stan.Conn
	logger *zap.Logger
}

func New(cluster string, clientID string, natsURL string, logger *zap.Logger) *Producer {
	sc, err := stan.Connect(cluster, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Fatal("Error connecting", zap.Error(err))
	}

	return &Producer{
		sc:     sc,
		logger: logger,
	}
}

type ProducerHandler struct {
	producer *Producer
	logger   *zap.Logger
	e        *echo.Echo
}

func NewProducerHandler(e *echo.Echo, producer *Producer, logger *zap.Logger) *ProducerHandler {
	handler := &ProducerHandler{
		producer: producer,
		logger:   logger,
		e:        e,
	}

	e.POST("/api/v1/producer/:msgCount", handler.Send)
	e.POST("/api/v1/producer/validate-fail/:msgCount", handler.SendFail)
	return handler
}

func (h *ProducerHandler) Send(c echo.Context) error {
	msgCount := c.Param("msgCount")
	count, err := strconv.Atoi(msgCount)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			h.logger.Error("Warning: error publishing msg id ", zap.String("nuid", ackedNuid), zap.Error(err))
		} else {
			h.logger.Info("Received ack for msg ", zap.String("nuid", ackedNuid))
		}
	}

	for i := 0; i < count; i++ {
		order := newFakeOrder(0, 2_000)
		or, _ := json.Marshal(order)
		_, err := h.producer.sc.PublishAsync("orders", or, ackHandler) // returns immediately
		if err != nil {
			h.logger.Error("Error publishing", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *ProducerHandler) SendFail(c echo.Context) error {
	msgCount := c.Param("msgCount")
	count, err := strconv.Atoi(msgCount)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			h.logger.Error("Warning: error publishing msg id ", zap.String("nuid", ackedNuid), zap.Error(err))
		} else {
			h.logger.Info("Received ack for msg ", zap.String("nuid", ackedNuid))
		}
	}

	for i := 0; i < count; i++ {
		order := newFakeOrder(-2_000, -1)
		or, _ := json.Marshal(order)
		_, err := h.producer.sc.PublishAsync("orders", or, ackHandler) // returns immediately
		if err != nil {
			h.logger.Error("Error publishing", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	}

	return c.NoContent(http.StatusOK)
}

func newFakeOrder(minPay int, maxPay int) model.Order {
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

	payment := newFakePayment(minPay, maxPay)
	payment.GoodsTotal = &totalSum

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

func newFakePayment(minPay int, maxPay int) model.Payment {
	faker := gofakeit.New(0)
	a := faker.IntRange(minPay, maxPay)
	d := faker.IntRange(minPay, maxPay)
	g := faker.IntRange(minPay, maxPay)
	c := faker.IntRange(minPay, maxPay)
	return model.Payment{
		Transaction:  faker.Word(),
		RequestID:    "",
		Currency:     faker.Currency().Short,
		Provider:     faker.Word(),
		Amount:       &a,
		PaymentDt:    faker.DateRange(time.Now().AddDate(0, -6, 0), time.Now()).Unix(),
		Bank:         faker.RandomString([]string{"alpha", "sberbank", "sovcombank"}),
		DeliveryCost: &d,
		GoodsTotal:   &g,
		CustomFee:    &c,
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
