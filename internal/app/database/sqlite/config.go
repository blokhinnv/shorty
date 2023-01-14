package database

import "github.com/blokhinnv/shorty/internal/app/server/config"

type SQLiteConfig struct {
	DBPath       string
	ClearOnStart bool
}

// Конструктор конфига SQLite на основе конфига сервера
func GetSQLiteConfig(cfg config.ServerConfig) SQLiteConfig {
	return SQLiteConfig{DBPath: cfg.SQLiteDBPath, ClearOnStart: cfg.SQLiteClearOnStart}
}
