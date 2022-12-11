package config

import (
	"regexp"

	"github.com/asaskevich/govalidator"
	"github.com/caarlos0/env/v6"
)

// Конфиг сервера
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"        valid:"url"`
	BaseURL       string `env:"BASE_URL"       envDefault:"http://localhost:8080" valid:"url"`
}

// Возвращает конфиг для сервера
func GetServerConfig() ServerConfig {
	cfg := ServerConfig{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	result, err := govalidator.ValidateStruct(cfg)
	if err != nil || !result {
		panic(err)
	}

	cfg.ServerAddress = regexp.MustCompile(`https?://`).ReplaceAllString(cfg.ServerAddress, "")
	return cfg
}
