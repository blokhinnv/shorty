package database

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
)

type PostgreConfig struct {
	// не буду указывать дефолтное значение DatabaseDSN, чтобы не хранить логин/пароль в коде
	DatabaseDSN  string
	ClearOnStart bool
}

// Конструктор конфига Postgre на основе конфига сервера
func GetPostgreConfig(cfg config.ServerConfig) PostgreConfig {
	return PostgreConfig{DatabaseDSN: cfg.PostgreDatabaseDSN, ClearOnStart: cfg.PostgreClearOnStart}
}
