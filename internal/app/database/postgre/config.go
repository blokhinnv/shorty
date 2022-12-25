package database

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/caarlos0/env/v6"
)

type PostgreConfig struct {
	DatabaseDSN  string `env:"DATABASE_DSN"      envDefault:"postgres://root:pwd@localhost:5432/root"`
	ClearOnStart bool   `env:"PG_CLEAR_ON_START" envDefault:"true"`
}

// Обновляет конфиг хранилища
func (cfg *PostgreConfig) UpdateFromFlags(flagCfg config.FlagConfig) {
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = flagCfg.DatabaseDSN
	}
}

// Конструктор конфига Postgre на основе переменных окружения
func GetPostgreConfig(flagCfg config.FlagConfig) PostgreConfig {
	var config PostgreConfig
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	config.UpdateFromFlags(flagCfg)
	return config
}
