package config

import (
	"regexp"

	"github.com/asaskevich/govalidator"
	"github.com/caarlos0/env/v6"
)

// Конфиг сервера
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"http://localhost:8080" valid:"url"`
	BaseURL       string `env:"BASE_URL"       envDefault:"http://localhost:8080" valid:"url"`
	SecretKey     string `env:"SECRET_KEY"     envDefault:"yandex-practicum"`
}

// Обновляет конфиг сервера на основе флагов
func (cfg *ServerConfig) UpdateFromFlags(flagCfg FlagConfig) {
	if flagCfg.BaseURL != "" {
		cfg.BaseURL = flagCfg.BaseURL
	}
	if flagCfg.ServerAddress != "" {
		cfg.ServerAddress = flagCfg.ServerAddress
	}
	if flagCfg.SecretKey != "" {
		cfg.SecretKey = flagCfg.SecretKey
	}
}

// Возвращает конфиг для сервера
func GetServerConfig(flagCfg FlagConfig) ServerConfig {
	cfg := ServerConfig{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	cfg.UpdateFromFlags(flagCfg)
	result, err := govalidator.ValidateStruct(cfg)
	if err != nil || !result {
		panic(err)
	}

	cfg.ServerAddress = regexp.MustCompile(`https?://`).ReplaceAllString(cfg.ServerAddress, "")
	return cfg
}
