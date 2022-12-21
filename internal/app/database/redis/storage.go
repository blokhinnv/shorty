package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-redis/redis"
)

type RedisStorage struct {
	shortToLongDB     *redis.Client
	keyExpirationTime time.Duration
}

// Конструктор нового хранилища URL
func NewRedisStorage(conf RedisConfig) (*RedisStorage, error) {
	shortToLong := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.ShortToLongDB,
	})
	_, err := shortToLong.Ping().Result()

	if err != nil {
		return nil, fmt.Errorf("can't access to DB: %v", err)
	}
	return &RedisStorage{shortToLong, conf.KeyExpirationHours}, nil
}

// Метод для добавления нового URL в БД
func (s *RedisStorage) AddURL(url, urlID string) {
	err := s.shortToLongDB.Set(urlID, url, s.keyExpirationTime).Err()
	if err != nil {
		panic(fmt.Sprintf("can't add record (%v, %v)into db\n", url, urlID))
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
}

// Возвращает URL по его ID в БД
func (s *RedisStorage) GetURLByID(urlID string) (string, error) {
	url, err := s.shortToLongDB.Get(urlID).Result()
	if errors.Is(err, redis.Nil) {
		return "", storage.ErrURLWasNotFound
	}
	return url, err
}

// Закрывает соединение со всеми БД Redis
func (s *RedisStorage) Close() {
	s.shortToLongDB.Close()
}
