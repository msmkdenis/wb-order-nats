package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
)

type Config struct {
	Address     string
	DatabaseURI string
	NatsCluster string
	NatsClient  string
	NatsURL     string
	NatsSubject string
	NatsQGroup  string
	NatsDurable string
}

func NewConfig() *Config {
	err := godotenv.Load("wborder.env")
	if err != nil {
		log.Info("Error loading .env file, using default values")
	}

	config := &Config{}
	config.Address = os.Getenv("RUN_ADDRESS")
	config.DatabaseURI = os.Getenv("DATABASE_URI")
	config.NatsCluster = os.Getenv("NATS_CLUSTER")
	config.NatsClient = os.Getenv("NATS_CLIENT")
	config.NatsURL = os.Getenv("NATS_URL")
	config.NatsSubject = os.Getenv("NATS_SUBJECT")
	config.NatsQGroup = os.Getenv("NATS_QGROUP")
	config.NatsDurable = os.Getenv("NATS_DURABLE")

	return config
}
