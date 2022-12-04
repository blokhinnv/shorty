package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/go-redis/redis"
)

const nTimesAddedKey = "__nAdded__"

type RedisStorage struct {
	longToShort       *redis.Client
	shortToLong       *redis.Client
	meta              *redis.Client
	keyExpirationTime time.Duration
}

// Конструктор нового хранилища URL
func NewRedisStorage(conf RedisConfig) (*RedisStorage, error) {
	longToShort := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.LongToShortDB,
	})
	_, err := longToShort.Ping().Result()

	if err != nil {
		return nil, fmt.Errorf("can't access to DB: %v", err)
	}

	shortToLong := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.ShortToLongDB,
	})
	_, err = shortToLong.Ping().Result()

	if err != nil {
		return nil, fmt.Errorf("can't access to DB: %v", err)
	}

	meta := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.MetaDB,
	})
	_, err = meta.Ping().Result()

	if err != nil {
		return nil, fmt.Errorf("can't access to DB: %v", err)
	}

	return &RedisStorage{longToShort, shortToLong, meta, conf.KeyExpirationHours}, nil
}

// Метод для добавления нового URL в БД
func (s *RedisStorage) AddURL(url, urlID string) {
	err := s.longToShort.Set(url, urlID, s.keyExpirationTime).Err()
	if err != nil {
		panic(fmt.Sprintf("can't add record (%v, %v)into db\n", url, urlID))
	}
	err = s.shortToLong.Set(urlID, url, s.keyExpirationTime).Err()
	if err != nil {
		panic(fmt.Sprintf("can't add record (%v, %v)into db\n", url, urlID))
	}

	s.meta.Incr(nTimesAddedKey)
}

// Возвращает URL по его ID в БД
func (s *RedisStorage) GetURLByID(urlID string) (string, error) {
	url, err := s.shortToLong.Get(urlID).Result()
	if errors.Is(err, redis.Nil) {
		return "", storage.ErrURLWasNotFound
	}
	return url, err
}

// Возвращает ID URL по его строковому представлению
func (s *RedisStorage) GetIDByURL(url string) (string, error) {
	urlID, err := s.longToShort.Get(url).Result()
	if errors.Is(err, redis.Nil) {
		return "", storage.ErrIDWasNotFound
	}
	return urlID, err
}

// Возвращает количество строк в таблице
func (s *RedisStorage) GetFreeUID() (int, error) {
	return s.meta.Get(nTimesAddedKey).Int()
}

// Закрывает соединение со всеми БД Redis
func (s *RedisStorage) Close() {
	s.longToShort.Close()
	s.shortToLong.Close()
	s.meta.Close()
}
