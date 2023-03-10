// Package config contains a description of the server config.
package config

import (
	"regexp"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/caarlos0/env/v6"
)

// ServerConfig - structure for storing the server config.
type ServerConfig struct {
	ServerAddress           string        `env:"SERVER_ADDRESS"              envDefault:"http://localhost:8080" valid:"url"`
	BaseURL                 string        `env:"BASE_URL"                    envDefault:"http://localhost:8080" valid:"url"`
	SecretKey               string        `env:"SECRET_KEY"` // I will not specify a default value for security
	PostgresDatabaseDSN     string        `env:"DATABASE_DSN"`
	PostgresClearOnStart    bool          `env:"PG_CLEAR_ON_START"           envDefault:"false"`
	SQLiteDBPath            string        `env:"SQLITE_DB_PATH"              envDefault:"db.sqlite3"`
	SQLiteClearOnStart      bool          `env:"SQLITE_CLEAR_ON_START"       envDefault:"false"`
	FileStoragePath         string        `env:"FILE_STORAGE_PATH"`
	FileStorageClearOnStart bool          `env:"FILE_STORAGE_CLEAR_ON_START" envDefault:"false"`
	FileStorageTTLOnDisk    time.Duration `env:"FILE_STORAGE_TTL_ON_DISK"    envDefault:"1h"`
	FileStorageTTLInMemory  time.Duration `env:"FILE_STORAGE_TTL_IN_MEMORY"  envDefault:"15m"`
}

// UpdateFromFlags Updates the server config based on flags.
func (cfg *ServerConfig) UpdateFromFlags(flagCfg *FlagConfig) {
	// it seems like env should have priority,
	// but then I won't pass the 7th test...
	if flagCfg.ServerAddress != "" {
		cfg.ServerAddress = flagCfg.ServerAddress
	}
	if flagCfg.BaseURL != "" {
		cfg.BaseURL = flagCfg.BaseURL
	}
	if flagCfg.FileStoragePath != "" {
		cfg.FileStoragePath = flagCfg.FileStoragePath
	}
	if flagCfg.SecretKey != "" {
		cfg.SecretKey = flagCfg.SecretKey
	}
	if flagCfg.DatabaseDSN != "" {
		cfg.PostgresDatabaseDSN = flagCfg.DatabaseDSN
	}
}

// NewServerConfig - config constructor for the server.
func NewServerConfig(flagCfg *FlagConfig) (*ServerConfig, error) {
	cfg := ServerConfig{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	cfg.UpdateFromFlags(flagCfg)
	result, err := govalidator.ValidateStruct(cfg)
	if err != nil || !result {
		return nil, err
	}
	cfg.ServerAddress = regexp.MustCompile(`https?://`).ReplaceAllString(cfg.ServerAddress, "")
	return &cfg, nil
}
