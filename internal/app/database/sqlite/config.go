package database

import (
	"github.com/caarlos0/env/v6"
)

type SQLiteConfig struct {
	DBPath string `env:"SQLITE_DB_PATH" envDefault:"db.sqlite3"`
}

// Конструктор конфига SQLite на основе переменных окружения
func GetSQLiteConfig() SQLiteConfig {
	var config SQLiteConfig
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	return config
}