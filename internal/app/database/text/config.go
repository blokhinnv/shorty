package text

import (
	"time"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// Конфиг текстового хранилища
type TextStorageConfig struct {
	FileStoragePath string
	ClearOnStart    bool
	TTLOnDisk       time.Duration
	TTLInMemory     time.Duration
}

// Конструктор конфига текстового хранилища на основе конфига сервера
func GetTextStorageConfig(cfg *config.ServerConfig) *TextStorageConfig {
	return &TextStorageConfig{
		FileStoragePath: cfg.FileStoragePath,
		ClearOnStart:    cfg.FileStorageClearOnStart,
		TTLOnDisk:       cfg.FileStorageTTLOnDisk,
		TTLInMemory:     cfg.FileStorageTTLInMemory,
	}
}
