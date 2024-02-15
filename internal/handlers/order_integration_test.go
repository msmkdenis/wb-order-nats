package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/app/fakeproducer"
	"github.com/msmkdenis/wb-order-nats/internal/cache/memory"
	"github.com/msmkdenis/wb-order-nats/internal/config"
	"github.com/msmkdenis/wb-order-nats/internal/consumer"
	"github.com/msmkdenis/wb-order-nats/internal/metrics"
	"github.com/msmkdenis/wb-order-nats/internal/middleware"
	"github.com/msmkdenis/wb-order-nats/internal/model"
	"github.com/msmkdenis/wb-order-nats/internal/repository"
	"github.com/msmkdenis/wb-order-nats/internal/service"
	"github.com/msmkdenis/wb-order-nats/internal/storage/db"
)

var cfgMock = &config.Config{}

type IntegrationTestSuite struct {
	suite.Suite
	orderHandler           *OrderHandler
	statisticsHandler      *StatisticsHandler
	producerHandler        *fakeproducer.ProducerHandler
	orderService           *service.OrderUseCase
	orderRepository        *repository.OrderRepository
	cache                  *memory.Cache
	natsClient             *consumer.NatsClient
	echo                   *echo.Echo
	postgresContainer      testcontainers.Container
	natsStreamingContainer testcontainers.Container
	pool                   *db.PostgresPool
	endpoint               string
	natsPort               nat.Port
	natsHost               string
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Error("Unable to initialize zap logger", zap.Error(err))
	}

	s.postgresContainer, s.pool, err = setupTestDatabase(logger)
	if err != nil {
		logger.Error("Unable to setup test database", zap.Error(err))
	}

	s.orderRepository = repository.NewOrderRepository(s.pool, logger)
	s.cache = memory.NewCache(logger)
	s.orderService = service.NewOrderUseCase(s.orderRepository, s.cache, logger)
	err = s.orderService.RestoreCache()
	if err != nil {
		logger.Error("failed to restore cache", zap.Error(err))
	}

	s.natsStreamingContainer, err = setupTestNatsStreaming(logger)
	if err != nil {
		logger.Error("Unable to setup test database", zap.Error(err))
	}

	s.natsPort, err = s.natsStreamingContainer.MappedPort(context.Background(), "4222")
	if err != nil {
		logger.Error("Unable to get nats port", zap.Error(err))
	}
	s.natsHost, err = s.natsStreamingContainer.Host(context.Background())
	if err != nil {
		logger.Error("Unable to get nats port", zap.Error(err))
	}

	statService := metrics.NewMessageStatsUseCase(logger)
	go statService.ProcessedMessagesRun(context.Background())

	s.natsClient, err = consumer.NewNatsClient("test-cluster", "test-consumer",
		fmt.Sprintf("http://%s:%d", s.natsHost, s.natsPort.Int()), s.orderService, statService, logger)
	if err != nil {
		logger.Error("failed to connect to nats-streaming", zap.Error(err))
	}

	if s.natsClient != nil {
		err = s.natsClient.OrderProcessingRun()
		if err != nil {
			logger.Error("failed to run order processing", zap.Error(err))
		}
	}

	cacheMiddleware := middleware.NewCacheMiddleware(s.cache, logger)

	s.echo = echo.New()

	s.orderHandler = NewOrderHandler(s.echo, s.orderService, cacheMiddleware, logger)
	s.statisticsHandler = NewStatisticsHandler(s.echo, statService, logger)

	producer := fakeproducer.New("test-cluster", "test-sender", fmt.Sprintf("http://%s:%d", s.natsHost, s.natsPort.Int()), logger)
	s.producerHandler = fakeproducer.NewProducerHandler(s.echo, producer, logger)
}

func (s *IntegrationTestSuite) TestProcessedMessagesCount() {
	producerReq := httptest.NewRequest(http.MethodPost, "/api/v1/producer/", nil)
	producerRec := httptest.NewRecorder()
	cProducer := s.echo.NewContext(producerReq, producerRec)
	cProducer.SetPath("/:msgCount")
	cProducer.SetParamNames("msgCount")
	cProducer.SetParamValues("1000")
	err := s.producerHandler.Send(cProducer)
	assert.NoError(s.T(), err)

	orderAllReq := httptest.NewRequest(http.MethodGet, "/api/v1/order/", nil)
	orderAllRec := httptest.NewRecorder()
	cAllOrder := s.echo.NewContext(orderAllReq, orderAllRec)

	statReq := httptest.NewRequest(http.MethodGet, "/api/v1/stats/counts", nil)
	statRec := httptest.NewRecorder()
	cStat := s.echo.NewContext(statReq, statRec)

	assert.Eventually(s.T(), func() bool {
		err := s.statisticsHandler.GetStatsCount(cStat)
		assert.NoError(s.T(), err)
		var stat StatCountsDTO
		err = json.Unmarshal(statRec.Body.Bytes(), &stat)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), 1_000, stat.Processed)

		err = s.orderHandler.FindAll(cAllOrder)
		assert.NoError(s.T(), err)
		var orders []model.Order
		err = json.Unmarshal(orderAllRec.Body.Bytes(), &orders)
		assert.NoError(s.T(), err)

		return len(orders) == stat.Processed-stat.Failed
	}, 10*time.Second, 2*time.Second)
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.postgresContainer.Terminate(context.Background())      //nolint:errcheck
	s.natsStreamingContainer.Terminate(context.Background()) //nolint:errcheck
}

func setupTestNatsStreaming(logger *zap.Logger) (testcontainers.Container, error) {
	containerReq := testcontainers.ContainerRequest{
		Image:        "nats-streaming:0.25.5-alpine",
		ExposedPorts: []string{"4222/tcp"},
		WaitingFor:   wait.ForListeningPort("4222/tcp"),
	}

	natsContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		logger.Info("error", zap.Error(err))
		return nil, err
	}

	return natsContainer, nil
}

func setupTestDatabase(logger *zap.Logger) (testcontainers.Container, *db.PostgresPool, error) {
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "wb-order-test",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
		},
	}
	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return nil, nil, err
	}

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		return nil, nil, err
	}
	host, err := dbContainer.Host(context.Background())
	if err != nil {
		return nil, nil, err
	}

	connection := fmt.Sprintf("user=postgres password=postgres host=%s database=wb-order-test sslmode=disable port=%d", host, port.Int())

	pool := initPostgresPool(context.Background(), connection, logger)

	return dbContainer, pool, err
}

func initPostgresPool(ctx context.Context, uri string, logger *zap.Logger) *db.PostgresPool {
	postgresPool, err := db.NewPostgresPool(ctx, uri, logger)
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
	}

	migrations, err := db.NewMigrations(uri, logger)
	if err != nil {
		logger.Fatal("Unable to create migrations", zap.Error(err))
	}

	err = migrations.MigrateUp()
	if err != nil {
		logger.Fatal("Unable to up migrations", zap.Error(err))
	}

	logger.Info("Connected to database", zap.String("DSN", uri))
	return postgresPool
}
