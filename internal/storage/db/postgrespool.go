package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/pkg/apperr"
)

// PostgresPool represents PostgreSQL connection pool.
type PostgresPool struct {
	DB     *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresPool returns a new instance of PostgresPool with pool of connections.
func NewPostgresPool(ctx context.Context, connection string, logger *zap.Logger) (*PostgresPool, error) {
	dbPool, err := pgxpool.New(ctx, connection)
	if err != nil {
		return nil, apperr.NewValueError(fmt.Sprintf("Unable to connect to database with connection %s", connection), apperr.Caller(), err)
	}
	logger.Info("Successful connection", zap.String("database", dbPool.Config().ConnConfig.Database))

	err = dbPool.Ping(ctx)
	if err != nil {
		return nil, apperr.NewValueError("Unable to ping database", apperr.Caller(), err)
	}
	logger.Info("Successful ping", zap.String("database", dbPool.Config().ConnConfig.Database))

	return &PostgresPool{
		DB:     dbPool,
		logger: logger,
	}, nil
}
