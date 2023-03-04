package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/blokhinnv/shorty/internal/app/database/postgres"
	"github.com/blokhinnv/shorty/internal/app/database/sqlite"
	"github.com/blokhinnv/shorty/internal/app/database/text"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/storage"
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
func inferStorageType(cfg *config.ServerConfig) int {
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
func NewDBStorage(cfg *config.ServerConfig) (storage.Storage, error) {
	storageType := inferStorageType(cfg)
	switch storageType {
	case SQLite:
		sqliteConfig := sqlite.GetSQLiteConfig(cfg)
		log.Printf("Starting SQLiteStorage with config %+v\n", sqliteConfig)
		return sqlite.NewSQLiteStorage(sqliteConfig)
	case Postgres:
		postgresConfig := postgres.GetPostgresConfig(cfg)
		log.Printf("Starting PostgreStorage with config %+v\n", postgresConfig)
		return postgres.NewPostgresStorage(postgresConfig)
	case Text:
		textStorageConfig := text.GetTextStorageConfig(cfg)
		log.Printf("Starting TextStorage with config %+v\n", textStorageConfig)
		return text.NewTextStorage(textStorageConfig)
	}
	return nil, fmt.Errorf("unknown storage type %v", storageType)
}
