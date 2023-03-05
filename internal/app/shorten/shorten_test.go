package shorten

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	log "github.com/sirupsen/logrus"
)

// BenchmarkToShortenBase - бенчмарк для toShortenBase.
func BenchmarkToShortenBase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toShortenBase(rand.Uint64())
	}
}

// BenchmarkGetShortURL - бенчмарк для GetShortURL.
func BenchmarkGetShortURL(b *testing.B) {
	b.Run("SQLite", func(b *testing.B) {
		cfg := &config.ServerConfig{
			ServerAddress:      "http://localhost:8080",
			BaseURL:            "http://localhost:8080",
			SecretKey:          "yandex-practicum",
			SQLiteDBPath:       "db_test.sqlite3",
			SQLiteClearOnStart: true,
		}
		s, err := database.NewDBStorage(cfg)
		if err != nil {
			log.Errorf("Can't run benchmarks for SQLite: %v", err.Error())
			return
		}
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL(s, "https://practicum.yandex.ru/", 1234, cfg.BaseURL)
		}
		b.StopTimer()
		os.Remove(cfg.SQLiteDBPath)
	})

	b.Run("Postgres", func(b *testing.B) {
		cfg := &config.ServerConfig{
			ServerAddress:        "http://localhost:8080",
			BaseURL:              "http://localhost:8080",
			SecretKey:            "yandex-practicum",
			PostgresDatabaseDSN:  "postgres://root:pwd@localhost:5432/root",
			PostgresClearOnStart: true,
		}
		s, err := database.NewDBStorage(cfg)
		if err != nil {
			log.Errorf("Can't run benchmarks for Postgres: %v", err.Error())
			return
		}
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL(s, "https://practicum.yandex.ru/", 1234, cfg.BaseURL)
		}
	})
	b.Run("Text", func(b *testing.B) {
		cfg := &config.ServerConfig{
			ServerAddress:           "http://localhost:8080",
			BaseURL:                 "http://localhost:8080",
			SecretKey:               "yandex-practicum",
			FileStoragePath:         "db_test.jsonl",
			FileStorageClearOnStart: true,
			FileStorageTTLOnDisk:    30 * time.Minute,
			FileStorageTTLInMemory:  10 * time.Minute,
		}
		s, err := database.NewDBStorage(cfg)
		if err != nil {
			log.Errorf("Can't run benchmarks for TextStorage: %v", err.Error())
			return
		}
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL(s, "https://practicum.yandex.ru/", 1234, cfg.BaseURL)
		}
		b.StopTimer()
		os.Remove(cfg.FileStoragePath)
	})

}
