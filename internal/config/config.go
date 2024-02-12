package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
}

func NewConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", "localhost:7000", "Адрес и порт запуска сервиса")
	flag.StringVar(&config.DatabaseURI, "d", "user=postgres password=postgres host=localhost database=wb-order sslmode=disable", "Адрес подключения к базе данных")

	if err := env.Parse(config); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return config
}
