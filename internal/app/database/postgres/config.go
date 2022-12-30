package database

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
)

type PostgresConfig struct {
	// не буду указывать дефолтное значение DatabaseDSN, чтобы не хранить логин/пароль в коде
	DatabaseDSN  string
	ClearOnStart bool
}

// Конструктор конфига Postgres на основе конфига сервера
func GetPostgresConfig(cfg config.ServerConfig) PostgresConfig {
	return PostgresConfig{
		DatabaseDSN:  cfg.PostgresDatabaseDSN,
		ClearOnStart: cfg.PostgresClearOnStart,
	}
}