package postgres

import (
	"context"
	"testing"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// BenchmarkPostgresStorage - бенчмарки для основных методов хранилища.
func BenchmarkPostgresStorage(b *testing.B) {
	cfg := &PostgresConfig{
		DatabaseDSN:  "postgres://root:pwd@localhost:5432/root",
		ClearOnStart: true,
	}
	s, err := NewPostgresStorage(cfg)
	if err != nil {
		log.Errorf("Can't run benchmarks for postgres: %v", err.Error())
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
