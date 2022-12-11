package database

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type RedisConfig struct {
	Addr               string        `env:"REDIS_ADDR"      envDefault:"localhost:6379"`
	Password           string        `env:"REDIS_PASSWORD"  envDefault:""`
	ShortToLongDB      int           `env:"REDIS_STL_DB"    envDefault:"1"`
	KeyExpirationHours time.Duration `env:"REDIS_HOURS_TTL" envDefault:"1h"`
}

// Конструктор конфига Redis на основе переменных окружения
func GetRedisConfig() RedisConfig {
	var config RedisConfig
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	return config
}
