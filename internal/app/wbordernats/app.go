package wbordernats

import (
	"context"
	"errors"
	"github.com/msmkdenis/wb-order-nats/internal/metrics"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/internal/cache/memory"
	"github.com/msmkdenis/wb-order-nats/internal/config"
	"github.com/msmkdenis/wb-order-nats/internal/consumer"
	"github.com/msmkdenis/wb-order-nats/internal/handlers"
	"github.com/msmkdenis/wb-order-nats/internal/middleware"
	"github.com/msmkdenis/wb-order-nats/internal/repository"
	"github.com/msmkdenis/wb-order-nats/internal/service"
	"github.com/msmkdenis/wb-order-nats/internal/storage/db"
)

func Run(quitSignal chan os.Signal) {
	cfg := *config.NewConfig()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Unable to initialize zap logger", err)
	}

	postgresPool := initPostgresPool(&cfg, logger)

	orderRepository := repository.NewOrderRepository(postgresPool, logger)
	cache := memory.NewCache(logger)

	orderService := service.NewOrderUseCase(orderRepository, cache, logger)
	err = orderService.RestoreCache()
	if err != nil {
		logger.Error("failed to restore cache", zap.Error(err))
	}

	statService := metrics.NewMessageStatsUseCase(logger)
	go statService.ProcessedMessagesRun(context.Background())

	nats, err := consumer.NewNatsClient("test-cluster", "test-client", "http://127.0.0.1:4222/", orderService, statService, logger)
	if err != nil {
		logger.Error("failed to connect to nats-streaming", zap.Error(err))
	}

	if nats != nil {
		err = nats.OrderProcessingRun()
		if err != nil {
			logger.Error("failed to run order processing", zap.Error(err))
		}
	}

	requestLogger := middleware.InitRequestLogger(logger)
	cacheMiddleware := middleware.NewCacheMiddleware(cache, logger)

	e := echo.New()

	e.Use(requestLogger.RequestLogger())

	handlers.NewOrderHandler(e, orderService, cacheMiddleware, logger)
	handlers.NewStatisticsHandler(e, statService, logger)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		<-quitSignal

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		if errShutdown := e.Shutdown(shutdownCtx); errShutdown != nil {
			e.Logger.Fatal(errShutdown)
		}
		serverStopCtx()
	}()

	go func() {
		errStart := e.Start(cfg.Address)
		if errStart != nil && !errors.Is(errStart, http.ErrServerClosed) {
			log.Fatal(errStart)
		}
	}()

	<-serverCtx.Done()
}

func initPostgresPool(cfg *config.Config, logger *zap.Logger) *db.PostgresPool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	postgresPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURI, logger)
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
	}

	migrations, err := db.NewMigrations(cfg.DatabaseURI, logger)
	if err != nil {
		logger.Fatal("Unable to create migrations", zap.Error(err))
	}

	err = migrations.MigrateUp()
	if err != nil {
		logger.Fatal("Unable to up migrations", zap.Error(err))
	}

	logger.Info("Connected to database", zap.String("DSN", cfg.DatabaseURI))
	return postgresPool
}
