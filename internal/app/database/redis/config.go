package database

import (
	"time"

	"github.com/blokhinnv/shorty/internal/app/env"
)

type RedisConfig = struct {
	Addr               string
	Password           string
	ShortToLongDB      int
	KeyExpirationHours time.Duration
}

const (
	DefaultAddr     = "localhost:6379"
	DefaultPassword = ""
	DefaultSTLDB    = 2
	DefaultHoursTTL = 1
)

// Конструктор конфига Redis на основе переменных окружения
func GetRedisConfig() RedisConfig {
	addr := env.GetOrDefault("REDIS_ADDR", DefaultAddr)
	pwd := env.GetOrDefault("REDIS_PASSWORD", DefaultPassword)
	stlDB := env.GetOrDefaultInt("REDIS_STL_DB", DefaultSTLDB)
	ttl := env.GetOrDefaultInt("REDIS_HOURS_TTL", DefaultHoursTTL)

	return RedisConfig{
		Addr:               addr,
		Password:           pwd,
		ShortToLongDB:      stlDB,
		KeyExpirationHours: time.Duration(ttl) * time.Hour,
	}
}
