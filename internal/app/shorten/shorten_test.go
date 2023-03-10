package shorten

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/blokhinnv/shorty/internal/app/server/config"
)

// BenchmarkToShortenBase - benchmark for toShortenBase.
func BenchmarkToShortenBase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toShortenBase(rand.Uint64())
	}
}

// BenchmarkGetShortURL - benchmark for GetShortURL.
func BenchmarkGetShortURL(b *testing.B) {
	b.Run("SQLite", func(b *testing.B) {
		cfg := &config.ServerConfig{
			ServerAddress:      "http://localhost:8080",
			BaseURL:            "http://localhost:8080",
			SecretKey:          "yandex-practicum",
			SQLiteDBPath:       "db_test.sqlite3",
			SQLiteClearOnStart: true,
		}
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL("https://practicum.yandex.ru/", 1234, cfg.BaseURL)
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
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL("https://practicum.yandex.ru/", 1234, cfg.BaseURL)
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
		b.ResetTimer()
		for i := 0; i <= b.N; i++ {
			GetShortURL("https://practicum.yandex.ru/", 1234, cfg.BaseURL)
		}
		b.StopTimer()
		os.Remove(cfg.FileStoragePath)
	})

}

func TestGetShortURL(t *testing.T) {
	type args struct {
		url     string
		userID  uint32
		baseURL string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "not_url",
			args: args{
				url:     "@.@@",
				userID:  1,
				baseURL: "localhost:8080",
			},
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name: "ok",
			args: args{
				url:     "http://yandex.ru",
				userID:  1,
				baseURL: "localhost:8080",
			},
			want:    "3lmmrhti6j0e2",
			want1:   "localhost:8080/3lmmrhti6j0e2",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetShortURL(tt.args.url, tt.args.userID, tt.args.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetShortURL() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetShortURL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
