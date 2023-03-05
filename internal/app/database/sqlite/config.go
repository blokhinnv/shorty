// Пакет sqlite реализует хранилище на основе SQLite.
package sqlite

import "github.com/blokhinnv/shorty/internal/app/server/config"

// SQLiteConfig - конфиг для хранилища на основе SQLite.
type SQLiteConfig struct {
	DBPath       string
	ClearOnStart bool
}

// GetSQLiteConfig - Конструктор конфига SQLite на основе конфига сервера.
func GetSQLiteConfig(cfg *config.ServerConfig) *SQLiteConfig {
	return &SQLiteConfig{DBPath: cfg.SQLiteDBPath, ClearOnStart: cfg.SQLiteClearOnStart}
}
