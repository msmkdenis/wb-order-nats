// Package db contains the database migrations and interactions.
package db

import (
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/msmkdenis/wb-order-nats/pkg/apperr"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrations represents the go migrate instance.
type Migrations struct {
	migrations *migrate.Migrate
	logger     *zap.Logger
}

// NewMigrations creates a new Migrations instance.
//
// It takes a connection string and a logger as parameters and returns a
// pointer to Migrations and an error.
func NewMigrations(connection string, logger *zap.Logger) (*Migrations, error) {
	dbConfig, err := pgxpool.ParseConfig(connection)
	if err != nil {
		return nil, apperr.NewValueError("Unable to parse connection string", apperr.Caller(), err)
	}
	logger.Info("Successful db url parsing", zap.String("database", dbConfig.ConnConfig.Database))

	dbURL := makeDBURL(dbConfig, parseSSLMode(connection))

	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, apperr.NewValueError("Unable to create iofs driver", apperr.Caller(), err)
	}
	logger.Info("Successful connection", zap.String("database", dbConfig.ConnConfig.Database))

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, dbURL)
	if err != nil {
		return nil, apperr.NewValueError("Unable to create new migrations", apperr.Caller(), err)
	}
	logger.Info("Successful migrations")

	return &Migrations{
		migrations: migrations,
		logger:     logger,
	}, nil
}

// MigrateUp perform migrations up.
func (m *Migrations) MigrateUp() error {
	err := m.migrations.Up()
	if err != nil && err.Error() != "no change" {
		return apperr.NewValueError("Unable to up migrations", apperr.Caller(), err)
	}
	return nil
}

func makeDBURL(config *pgxpool.Config, sslMode string) string {
	var dbURL strings.Builder

	dbURL.WriteString("postgres://")
	dbURL.WriteString(config.ConnConfig.User)
	dbURL.WriteString(":")
	dbURL.WriteString(config.ConnConfig.Password)
	dbURL.WriteString("@")
	dbURL.WriteString(config.ConnConfig.Host)
	dbURL.WriteString(":")
	dbURL.WriteString(fmt.Sprint(config.ConnConfig.Port))
	dbURL.WriteString("/")
	dbURL.WriteString(config.ConnConfig.Database)
	dbURL.WriteString("?sslmode=")
	if config.ConnConfig.TLSConfig == nil {
		dbURL.WriteString("disable")
	} else {
		dbURL.WriteString(sslMode)
	}

	return dbURL.String()
}

func parseSSLMode(connection string) string {
	con := strings.Split(connection, " ")
	sslMode := ""
	for _, v := range con {
		pair := strings.Split(v, "=")
		if pair[0] == "sslmode" {
			sslMode = pair[1]
		}
	}

	return sslMode
}
