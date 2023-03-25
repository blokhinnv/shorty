// Пакет sqlite реализует хранилище на основе SQLite.
package sqlite

import "github.com/blokhinnv/shorty/internal/app/server/config"

// SQLiteConfig - config for storage based on SQLite.
type SQLiteConfig struct {
	DBPath       string
	ClearOnStart bool
}

// GetSQLiteConfig - SQLite config constructor based on server config.
func GetSQLiteConfig(cfg *config.ServerConfig) *SQLiteConfig {
	return &SQLiteConfig{DBPath: cfg.SQLiteDBPath, ClearOnStart: cfg.SQLiteClearOnStart}
}
