package database

import (
	"github.com/blokhinnv/shorty/internal/app/env"
)

type SQLiteConfig = struct {
	DBPath string
}

const (
	DefaultDBPath = "db.sqlite3"
)

// Конструктор конфига SQLite на основе переменных окружения
func GetSQLiteConfig() SQLiteConfig {
	dbPath := env.GetOrDefault("SQLITE_DB_PATH", DefaultDBPath)
	return SQLiteConfig{dbPath}
}
