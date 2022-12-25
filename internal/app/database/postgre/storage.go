// Пакет для создания БД - хранилища URL
package database

import (
	"context"
	"log"
	"os"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/jackc/pgx/v5"
)

const (
	selectByURLIDSQL  = "SELECT url, user_id FROM Url WHERE url_id = $1"
	selectByUserIDSQL = "SELECT url, url_id FROM Url WHERE user_id = $1"
	insertSQL         = "INSERT INTO Url(url, url_id, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, url_id) DO NOTHING;"
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
	_, err := s.conn.Exec(context.Background(), insertSQL, url, urlID, userID)
	if err != nil {
		log.Fatalf("Error while adding URL: %v", err)
		return err
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *PostgreStorage) GetURLByID(urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	// Получаем строки
	err := s.conn.QueryRow(context.Background(), selectByURLIDSQL, urlID).
		Scan(&rec.URL, &rec.UserID)
	// любая ошибка здесь (в т.ч. ErrNoRows) означает, что результат не найден
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	return rec, nil
}

func (s *PostgreStorage) GetURLsByUser(userID uint32) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.conn.Query(context.Background(), selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Проходим по всем записям методом rows.Next() до тех пор,
	// пока не пройдём все доступные результаты
	for rows.Next() {
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	// После цикла проверяем записи на потенциальные ошибки (разрыв
	// сетевого соединения с сервером базы данных в процессе получения результатов запроса)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, storage.ErrURLWasNotFound
	}
	return results, nil
}

// Закрывает соединение с SQLite
func (s *PostgreStorage) Close() {
	s.conn.Close(context.Background())
}

func (s *PostgreStorage) Ping() bool {
	return s.conn.Ping(context.Background()) == nil
}
