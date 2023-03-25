// Package database contains implementations of Storage based on various storage types.
package database

import (
	"fmt"

	"github.com/blokhinnv/shorty/internal/app/log"

	"github.com/blokhinnv/shorty/internal/app/database/postgres"
	"github.com/blokhinnv/shorty/internal/app/database/sqlite"
	"github.com/blokhinnv/shorty/internal/app/database/text"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Storage types.
const (
	Postgres = iota
	Text
	SQLite
)

// inferStorageType determines which storage type to use
// based on config.
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

// NewDBStorage - database-based storage constructor.
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
