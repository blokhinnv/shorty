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
	DefaultAddr     = "localhost:6379"
	DefaultPassword = ""
	DefaultLTSDB    = 1
	DefaultSTLDB    = 2
	DefaultMetaDB   = 3
	DefaultHoursTTL = 1
)

// Конструктор конфига Redis на основе переменных окружения
func GetRedisConfig() RedisConfig {
	addr := env.GetOrDefault("REDIS_ADDR", DefaultAddr)
	pwd := env.GetOrDefault("REDIS_PASSWORD", DefaultPassword)

	ltsDB := env.GetOrDefaultInt("REDIS_LTS_DB", DefaultLTSDB)
	stlDB := env.GetOrDefaultInt("REDIS_STL_DB", DefaultSTLDB)
	metaDB := env.GetOrDefaultInt("REDIS_META_DB", DefaultMetaDB)
	ttl := env.GetOrDefaultInt("REDIS_HOURS_TTL", DefaultHoursTTL)

	return RedisConfig{
		Addr:               addr,
		Password:           pwd,
		ShortToLongDB:      stlDB,
		LongToShortDB:      ltsDB,
		MetaDB:             metaDB,
		KeyExpirationHours: time.Duration(ttl) * time.Hour,
	}
}
