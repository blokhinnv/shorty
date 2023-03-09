package sqlite

import (
	"context"
	"testing"

	log "github.com/sirupsen/logrus"
)

// BenchmarkSQLiteStorage - бенчмарки для основных методов хранилища.
func BenchmarkSQLiteStorage(b *testing.B) {
	cfg := &SQLiteConfig{
		DBPath:       "db_test.sqlite3",
		ClearOnStart: true,
	}
	s, err := NewSQLiteStorage(cfg)
	if err != nil {
		log.Errorf("Can't run benchmarks for sqlite: %v", err.Error())
		return
	}
	ctx := context.Background()
	log.SetLevel(log.WarnLevel)
	b.ResetTimer()
	b.Run("AddURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.AddURL(ctx, "http://yandex.ru", "zxcvbn", 2)
		}
	})
	b.Run("GetURLByID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.GetURLByID(ctx, "zxcvbn")
		}
	})
	b.Run("GetURLsByUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.GetURLsByUser(ctx, 2)
		}
	})
	b.Run("AddURLBatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "zxcvbn"}, 2)
		}
	})
	b.Run("DeleteMany", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.DeleteMany(ctx, 2, []string{"zxcvbn"})
		}
	})
}
