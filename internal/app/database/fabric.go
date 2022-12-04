package database

import (
	"fmt"
	"os"

	redis "github.com/blokhinnv/shorty/internal/app/database/redis"
	sqlite "github.com/blokhinnv/shorty/internal/app/database/sqlite"
	storage "github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	SQLite = "sqlite"
	Redis  = "redis"
)

// Конструктор хранилища на основе БД
func NewDBStorage() storage.Storage {
	var storage storage.Storage
	var err error

	storageType := os.Getenv("STORAGE")
	switch storageType {
	case SQLite:
		sqliteConfig := sqlite.GetSQLiteConfig()
		storage, err = sqlite.NewSQLiteStorage(sqliteConfig)
	case Redis:
		redisConfig := redis.GetRedisConfig()
		storage, err = redis.NewRedisStorage(redisConfig)
	default:
		panic(fmt.Sprintf("unknown storage type %v", storageType))
	}
	if err != nil {
		panic("can't create a storage")
	}
	return storage
}
