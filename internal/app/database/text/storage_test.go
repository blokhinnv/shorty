package text

import (
	"context"
	"testing"
	"time"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/stretchr/testify/suite"
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

type TextSuite struct {
	suite.Suite
	textCfg *TextStorageConfig
}

func (suite *TextSuite) SetupSuite() {
	serverCfg := &config.ServerConfig{
		FileStoragePath:         "test.jsonl",
		FileStorageClearOnStart: true,
		FileStorageTTLOnDisk:    3 * time.Second,
		FileStorageTTLInMemory:  3 * time.Second,
	}
	suite.textCfg = GetTextStorageConfig(serverCfg)
}

func (suite *TextSuite) TearDownSuite() {
}

func (suite *TextSuite) TestAddURL() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestAddURLTwice() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)

	err = s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestGetURLByIDOk() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	rec, err := s.GetURLByID(ctx, "qwerty")
	suite.NoError(err)
	suite.Equal("http://yandex.ru", rec.URL)
	s.Close(ctx)
}

func (suite *TextSuite) TestGetURLByIDEmpty() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	_, err := s.GetURLByID(ctx, "qwerty")
	suite.Error(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestGetURLsByUserNotFound() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	_, err := s.GetURLsByUser(ctx, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestGetURLsByUserFound() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	res, err := s.GetURLsByUser(ctx, uint32(1))
	suite.NoError(err)
	suite.Equal("http://yandex.ru", res[0].URL)
	s.Close(ctx)
}

func (suite *TextSuite) TestBatchOK() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)

	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestBatchErr() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestDelete() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.DeleteMany(ctx, uint32(1), []string{"qwerty"})
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestPing() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	ping := s.Ping(ctx)
	suite.Equal(true, ping)
	s.Close(ctx)
}

func (suite *TextSuite) TestClear() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	err := s.Clear(ctx)
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *TextSuite) TestUpdate() {
	ctx := context.Background()
	s, _ := NewTextStorage(suite.textCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)
	time.Sleep(4 * time.Second)
	s.Close(ctx)
}

func TestTextSuite(t *testing.T) {
	suite.Run(t, new(TextSuite))
}
