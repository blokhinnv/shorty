package database

import (
	"fmt"
	"os"
	"strings"
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

// Конструктор конфига Redis на основе переменных окружения
func GetRedisConfig() RedisConfig {
	missing := make([]string, 0)
	requiredVars := []string{
		"REDIS_ADDR",
		"REDIS_HOURS_TTL",
		"REDIS_LTS_DB",
		"REDIS_STL_DB",
		"REDIS_META_DB",
	}
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}
	if len(missing) > 0 {
		panic(fmt.Sprintf("missing %s env variable", strings.Join(missing, ",")))
	}

	addr := os.Getenv("REDIS_ADDR")
	pwd := os.Getenv("REDIS_PASSWORD")

	ltsDB := env.VarToInt("REDIS_LTS_DB")
	stlDB := env.VarToInt("REDIS_STL_DB")
	metaDB := env.VarToInt("REDIS_META_DB")
	ttl := env.VarToInt("REDIS_HOURS_TTL")

	return RedisConfig{
		Addr:               addr,
		Password:           pwd,
		ShortToLongDB:      stlDB,
		LongToShortDB:      ltsDB,
		MetaDB:             metaDB,
		KeyExpirationHours: time.Duration(ttl) * time.Hour,
	}
}
