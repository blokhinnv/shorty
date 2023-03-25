package postgres

import (
	"context"
	"testing"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/stretchr/testify/suite"
)

// BenchmarkPostgresStorage - benchmarks for the main storage methods.
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

type PostgresSuite struct {
	suite.Suite
	pgCfg *PostgresConfig
}

var serverCfg = &config.ServerConfig{
	PostgresDatabaseDSN:  "postgres://root:pwd@localhost:5432/root",
	PostgresClearOnStart: true,
}

var pgCfg = GetPostgresConfig(serverCfg)

func (suite *PostgresSuite) TestAddURL() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestAddURLTwice() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	err := s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.NoError(err)

	err = s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestGetURLByIDOk() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	rec, err := s.GetURLByID(ctx, "qwerty")
	suite.NoError(err)
	suite.Equal("http://yandex.ru", rec.URL)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestGetURLByIDEmpty() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	_, err := s.GetURLByID(ctx, "qwerty")
	suite.Error(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestGetURLsByUserNotFound() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	_, err := s.GetURLsByUser(ctx, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestGetURLsByUserFound() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	res, err := s.GetURLsByUser(ctx, uint32(1))
	suite.NoError(err)
	suite.Equal("http://yandex.ru", res[0].URL)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestBatchOK() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)

	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestBatchErr() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.AddURLBatch(ctx, map[string]string{"http://yandex.ru": "qwerty"}, uint32(1))
	suite.Error(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestDelete() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	s.AddURL(ctx, "http://yandex.ru", "qwerty", uint32(1))
	err := s.DeleteMany(ctx, uint32(1), []string{"qwerty"})
	suite.NoError(err)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestPing() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	ping := s.Ping(ctx)
	suite.Equal(true, ping)
	s.Close(ctx)
}

func (suite *PostgresSuite) TestClear() {
	ctx := context.Background()
	s, _ := NewPostgresStorage(pgCfg)
	err := s.Clear(ctx)
	suite.NoError(err)
	s.Close(ctx)
}

func TestPostgresSuite(t *testing.T) {
	ps := new(PostgresSuite)
	_, err := NewPostgresStorage(pgCfg)
	if err != nil {
		t.Skip("Skipping test, reason: connection to postgres cannot be established")
	}
	suite.Run(t, ps)
}
