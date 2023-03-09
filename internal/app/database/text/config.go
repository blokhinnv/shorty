package text

import (
	"time"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// TextStorageConfig - конфиг текстового хранилища.
type TextStorageConfig struct {
	FileStoragePath string
	ClearOnStart    bool
	TTLOnDisk       time.Duration
	TTLInMemory     time.Duration
}

// GetTextStorageConfig - конструктор конфига текстового хранилища на основе конфига сервера.
func GetTextStorageConfig(cfg *config.ServerConfig) *TextStorageConfig {
	return &TextStorageConfig{
		FileStoragePath: cfg.FileStoragePath,
		ClearOnStart:    cfg.FileStorageClearOnStart,
		TTLOnDisk:       cfg.FileStorageTTLOnDisk,
		TTLInMemory:     cfg.FileStorageTTLInMemory,
	}
}
