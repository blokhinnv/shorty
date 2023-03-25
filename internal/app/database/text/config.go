package text

import (
	"time"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// TextStorageConfig - text storage config.
type TextStorageConfig struct {
	FileStoragePath string
	ClearOnStart    bool
	TTLOnDisk       time.Duration
	TTLInMemory     time.Duration
}

// GetTextStorageConfig - text storage config constructor based on server config.
func GetTextStorageConfig(cfg *config.ServerConfig) *TextStorageConfig {
	return &TextStorageConfig{
		FileStoragePath: cfg.FileStoragePath,
		ClearOnStart:    cfg.FileStorageClearOnStart,
		TTLOnDisk:       cfg.FileStorageTTLOnDisk,
		TTLInMemory:     cfg.FileStorageTTLInMemory,
	}
}
