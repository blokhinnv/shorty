package database

import (
	"fmt"
	"log"
	"os"

	redis "github.com/blokhinnv/shorty/internal/app/database/redis"
	sqlite "github.com/blokhinnv/shorty/internal/app/database/sqlite"
	text "github.com/blokhinnv/shorty/internal/app/database/text"
	"github.com/blokhinnv/shorty/internal/app/env"
	storage "github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	SQLite = "sqlite"
	Redis  = "redis"
	Text   = "text"
)

const (
	StorageEnvVar      = "STORAGE"
	FileStoragePathVar = "FILE_STORAGE_PATH"
)

// Конструктор хранилища на основе БД
func NewDBStorage() storage.Storage {
	var storage storage.Storage
	var err error

	// я бы лучше оставил switch на 3 разных хранилища
	// но в задании четко сказано, что если есть FILE_STORAGE_PATH,
	// то надо использовать текстовые файлы для хранения
	if os.Getenv(FileStoragePathVar) != "" {
		textStorageConfig := text.GetTextStorageConfig()
		log.Printf("Starting TextStorage with config %+v\n", textStorageConfig)
		storage, err = text.NewTextStorage(textStorageConfig)
	} else {
		storageType := env.GetOrDefault(StorageEnvVar, SQLite)
		switch storageType {
		case SQLite:
			sqliteConfig := sqlite.GetSQLiteConfig()
			log.Printf("Starting SQLiteStorage with config %+v\n", sqliteConfig)
			storage, err = sqlite.NewSQLiteStorage(sqliteConfig)
		case Redis:
			redisConfig := redis.GetRedisConfig()
			log.Printf("Starting RedisStorage with config %+v\n", redisConfig)
			storage, err = redis.NewRedisStorage(redisConfig)
		case Text:
			textStorageConfig := text.GetTextStorageConfig()
			log.Printf("Starting TextStorage with config %+v\n", textStorageConfig)
			storage, err = text.NewTextStorage(textStorageConfig)
		default:
			panic(fmt.Sprintf("unknown storage type %v", storageType))
		}
	}
	if err != nil {
		panic(fmt.Sprintf("can't create a storage: %v", err.Error()))
	}
	return storage
}
