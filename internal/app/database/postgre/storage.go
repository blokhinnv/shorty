// Пакет для создания БД - хранилища URL

// Пока не понимаю, какое должно быть хранилище
// создаю структуру UrlStorage, которая реализует
// интерфейс Storage. Если БД использовать нельзя,
// создам что-то другое, реализующее тот же интерфейс
package database

import (
	"context"
	"log"
	"os"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/jackc/pgx/v5"
)

const (
	selectByURLIDSQL  = "SELECT url, user_id FROM Url WHERE url_id = ?"
	selectByUserIDSQL = "SELECT url, url_id FROM Url WHERE user_id = ?"
	insertSQL         = "INSERT OR REPLACE INTO Url(url, url_id, user_id) VALUES (?, ?, ?)"
)

type PostgreStorage struct {
	conn *pgx.Conn
}

// Конструктор нового хранилища URL
func NewPostgreStorage(conf PostgreConfig) (*PostgreStorage, error) {
	conn, err := pgx.Connect(context.Background(), conf.DatabaseDSN)
	if err != nil {
		log.Fatalf("can't access to DB %s: %v\n", conf.DatabaseDSN, err)
		os.Exit(1)
	}
	InitDB(conn, conf.ClearOnStart)
	return &PostgreStorage{conn}, nil
}

// Метод для добавления нового URL в БД
func (s *PostgreStorage) AddURL(url, urlID string, userID uint32) error {
	return nil
}

// Возвращает URL по его ID в БД
func (s *PostgreStorage) GetURLByID(urlID string) (storage.Record, error) {
	return storage.Record{}, nil
}

func (s *PostgreStorage) GetURLsByUser(userID uint32) ([]storage.Record, error) {
	return nil, nil
}

// Закрывает соединение с SQLite
func (s *PostgreStorage) Close() {
	s.conn.Close(context.Background())
}

func (s *PostgreStorage) Ping() bool {
	return s.conn.Ping(context.Background()) == nil
}
