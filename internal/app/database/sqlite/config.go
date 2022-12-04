package database

import (
	"github.com/blokhinnv/shorty/internal/app/env"
)

type SQLiteConfig = struct {
	DBPath string
}

const (
	DEFAULT_DB_PATH = "db.sqlite3"
)

// Конструктор конфига SQLite на основе переменных окружения
func GetSQLiteConfig() SQLiteConfig {
	dbPath := env.GetOrDefault("SQLITE_DB_PATH", DEFAULT_DB_PATH)
	return SQLiteConfig{dbPath}
}
