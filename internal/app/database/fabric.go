package database

import (
	"fmt"
	"log"

	postgres "github.com/blokhinnv/shorty/internal/app/database/postgres"
	sqlite "github.com/blokhinnv/shorty/internal/app/database/sqlite"
	text "github.com/blokhinnv/shorty/internal/app/database/text"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	storage "github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	Postgres = iota
	Text
	SQLite
)

const (
	StorageEnvVar      = "STORAGE"
	FileStoragePathVar = "FILE_STORAGE_PATH"
)

// Определяет, какой тип хранилища надо использовать
// на основе конфига
func inferStorageType(cfg config.ServerConfig) int {
	switch {
	case cfg.PostgresDatabaseDSN != "":
		return Postgres
	case cfg.FileStoragePath != "":
		return Text
	default:
		return SQLite
	}
}

// Конструктор хранилища на основе БД
func NewDBStorage(cfg config.ServerConfig) storage.Storage {
	var storage storage.Storage
	var err error

	storageType := inferStorageType(cfg)
	switch storageType {
	case SQLite:
		sqliteConfig := sqlite.GetSQLiteConfig(cfg)
		log.Printf("Starting SQLiteStorage with config %+v\n", sqliteConfig)
		storage, err = sqlite.NewSQLiteStorage(sqliteConfig)
	case Postgres:
		postgresConfig := postgres.GetPostgresConfig(cfg)
		log.Printf("Starting PostgreStorage with config %+v\n", postgresConfig)
		storage, err = postgres.NewPostgresStorage(postgresConfig)
	case Text:
		textStorageConfig := text.GetTextStorageConfig(cfg)
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
