package database

import (
	"fmt"
	"log"
	"os"

	sqlite "github.com/blokhinnv/shorty/internal/app/database/sqlite"
	text "github.com/blokhinnv/shorty/internal/app/database/text"
	"github.com/blokhinnv/shorty/internal/app/env"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	storage "github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	SQLite = "sqlite"
	Text   = "text"
)

const (
	StorageEnvVar      = "STORAGE"
	FileStoragePathVar = "FILE_STORAGE_PATH"
)

// Конструктор хранилища на основе БД
func NewDBStorage(flagCfg config.FlagConfig) storage.Storage {
	var storage storage.Storage
	var err error

	storageType := env.GetOrDefault(StorageEnvVar, SQLite)
	// При отсутствии переменной окружения или при её пустом значении
	// вернитесь к хранению сокращённых URL в памяти.
	if os.Getenv(FileStoragePathVar) != "" || flagCfg.FileStoragePath != "" {
		storageType = Text
	}
	switch storageType {
	case SQLite:
		sqliteConfig := sqlite.GetSQLiteConfig()
		log.Printf("Starting SQLiteStorage with config %+v\n", sqliteConfig)
		storage, err = sqlite.NewSQLiteStorage(sqliteConfig)
	case Text:
		textStorageConfig := text.GetTextStorageConfig(flagCfg)
		log.Printf("Starting TextStorage with config %+v\n", textStorageConfig)
		storage, err = text.NewTextStorage(textStorageConfig)
	default:
		panic(fmt.Sprintf("unknown storage type %v", storageType))
	}
	if err != nil {
		panic(fmt.Sprintf("can't create a storage: %v", err.Error()))
	}
	return storage
}
