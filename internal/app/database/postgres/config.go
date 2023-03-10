// Package postgres implements Postgres-based storage.
package postgres

import (
	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// PostgresConfig - config for Postgres-based storage.
type PostgresConfig struct {
	// I will not specify the default DatabaseDSN value,
	// not to store login/password in code
	DatabaseDSN  string
	ClearOnStart bool
}

// GetPostgresConfig - Postgres config constructor based on server config
func GetPostgresConfig(cfg *config.ServerConfig) *PostgresConfig {
	return &PostgresConfig{
		DatabaseDSN:  cfg.PostgresDatabaseDSN,
		ClearOnStart: cfg.PostgresClearOnStart,
	}
}
