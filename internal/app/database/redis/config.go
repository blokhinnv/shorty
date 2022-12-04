package database

import (
	"time"

	"github.com/blokhinnv/shorty/internal/app/env"
)

type RedisConfig = struct {
	Addr               string
	Password           string
	ShortToLongDB      int
	LongToShortDB      int
	MetaDB             int
	KeyExpirationHours time.Duration
}

const (
	DEFAULT_ADDR      = "localhost:6379"
	DEFAULT_PASSWORD  = ""
	DEFAULT_LTS_DB    = 1
	DEFAULT_STL_DB    = 2
	DEFAULT_META_DB   = 3
	DEFAULT_HOURS_TTL = 1
)

// Конструктор конфига Redis на основе переменных окружения
func GetRedisConfig() RedisConfig {
	addr := env.GetOrDefault("REDIS_ADDR", DEFAULT_ADDR)
	pwd := env.GetOrDefault("REDIS_PASSWORD", DEFAULT_PASSWORD)

	ltsDB := env.GetOrDefaultInt("REDIS_LTS_DB", DEFAULT_LTS_DB)
	stlDB := env.GetOrDefaultInt("REDIS_STL_DB", DEFAULT_STL_DB)
	metaDB := env.GetOrDefaultInt("REDIS_META_DB", DEFAULT_META_DB)
	ttl := env.GetOrDefaultInt("REDIS_HOURS_TTL", DEFAULT_HOURS_TTL)

	return RedisConfig{
		Addr:               addr,
		Password:           pwd,
		ShortToLongDB:      stlDB,
		LongToShortDB:      ltsDB,
		MetaDB:             metaDB,
		KeyExpirationHours: time.Duration(ttl) * time.Hour,
	}
}
