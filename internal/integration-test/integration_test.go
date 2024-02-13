package integration_test

import (
	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/magiconair/properties/assert"
	"github.com/msmkdenis/wb-order-nats/internal/cache/memory"
	"github.com/msmkdenis/wb-order-nats/internal/config"
	"github.com/msmkdenis/wb-order-nats/internal/consumer"
	"github.com/msmkdenis/wb-order-nats/internal/handlers"
	"github.com/msmkdenis/wb-order-nats/internal/repository"
	"github.com/msmkdenis/wb-order-nats/internal/service"
	"github.com/msmkdenis/wb-order-nats/internal/storage/db"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"testing"
)

var cfgMock = &config.Config{}

type IntegrationTestSuite struct {
	suite.Suite
	orderHandler           *handlers.OrderHandler
	orderService           *service.OrderUseCase
	orderRepository        *repository.OrderRepository
	cache                  *memory.Cache
	natsClient             *consumer.NatsClient
	echo                   *echo.Echo
	postgresContainer      testcontainers.Container
	natsStreamingContainer testcontainers.Container
	pool                   *db.PostgresPool
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

	//natsUrl := fmt.Sprintf("nats://127.0.0.1:%s", s.natsStreamingContainer.GetPort("4222/tcp"))
	//s.natsClient, err = consumer.NewNatsClient("test-cluster", "test-client", "http://127.0.0.1:4222/", s.orderService, logger)
	//if err != nil {
	//	logger.Error("failed to connect to nats-streaming", zap.Error(err))
	//}
	//
	//s.urlService = service.NewURLService(s.urlRepository, logger)
	//s.echo = echo.New()
	//s.endpoint, err = s.container.Endpoint(context.Background(), "http")
	//if err != nil {
	//	logger.Error("Unable to get endpoint", zap.Error(err))
	//}
	//s.urlHandler = handlers.NewURLHandler(s.echo, s.urlService, s.endpoint, jwtManager, logger)
}

func (s *IntegrationTestSuite) TestAB() {
	a := 2 + 2
	assert.Equal(s.T(), 4, a)
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

	fmt.Println("===========================")
	host, _ := natsContainer.Host(context.Background())
	fmt.Println(host)
	fmt.Println("===========================")

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
