package database

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// Конфиг текстового хранилища
type TextStorageConfig = struct {
	FileStoragePath string        `env:"FILE_STORAGE_PATH"           envDefault:"db.jsonl"`
	ClearOnStart    bool          `env:"FILE_STORAGE_CLEAR_ON_START" envDefault:"false"`
	TTLOnDisk       time.Duration `env:"FILE_STORAGE_TTL_ON_DISK"    envDefault:"1h"`
	TTLInMemory     time.Duration `env:"FILE_STORAGE_TTL_IN_MEMORY"  envDefault:"15m"`
}

// Конструктор конфига текстового хранилища на основе переменных окружения
func GetTextStorageConfig() TextStorageConfig {
	var config TextStorageConfig
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	return config
}
