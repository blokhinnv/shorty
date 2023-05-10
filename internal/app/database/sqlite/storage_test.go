package sqlite

import (
	"context"
	"testing"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/stretchr/testify/suite"
)

// BenchmarkSQLiteStorage - benchmarks for the main storage methods.
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

type SQLiteSuite struct {
	suite.Suite
	sqliteCfg *SQLiteConfig
}

func (suite *SQLiteSuite) SetupSuite() {
	serverCfg := &config.ServerConfig{
		SQLiteDBPath:       "test.sqlite3",
		SQLiteClearOnStart: true,
	}
	suite.sqliteCfg = GetSQLiteConfig(serverCfg)
}

func (suite *SQLiteSuite) TearDownSuite() {
}

func (suite *SQLiteSuite) TestAddURL() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestAddURLTwice() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)

	err = s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestGetURLByIDOk() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	rec, err := s.GetURLByID(ctx, "qwerty")
	suite.NoError(err)
	suite.Equal("http://yandex.ru", rec.URL)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestGetURLByIDEmpty() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	_, err := s.GetURLByID(ctx, "qwerty")
	suite.Error(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestGetURLsByUserNotFound() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	_, err := s.GetURLsByUser(ctx, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestGetURLsByUserFound() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	res, err := s.GetURLsByUser(ctx, uint32(1))
	suite.NoError(err)
	suite.Equal("http://yandex.ru", res[0].URL)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestBatchOK() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)

	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestBatchErr() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestDelete() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.DeleteMany(ctx, uint32(1), []string{"qwerty"})
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestPing() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	ping := s.Ping(ctx)
	suite.Equal(true, ping)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestClear() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	err := s.Clear(ctx)
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *SQLiteSuite) TestStats() {
	ctx := context.Background()
	s, _ := NewSQLiteStorage(suite.sqliteCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	urls, users, err := s.GetStats(ctx)
	suite.Equal(1, urls)
	suite.Equal(1, users)
	suite.NoError(err)
	s.Close(ctx)
}

func TestSQLiteSuite(t *testing.T) {
	suite.Run(t, new(SQLiteSuite))
}
