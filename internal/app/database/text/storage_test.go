package text

import (
	"context"
	"testing"
	"time"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// BenchmarkTextStorage - бенчмарки для основных методов хранилища.
func BenchmarkTextStorage(b *testing.B) {
	cfg := &TextStorageConfig{
		FileStoragePath: "db_test.jsonl",
		ClearOnStart:    true,
		TTLOnDisk:       30 * time.Minute,
		TTLInMemory:     10 * time.Minute,
	}
	s, err := NewTextStorage(cfg)
	if err != nil {
		panic(err)
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
