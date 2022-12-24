// Пакет для создания БД - хранилища URL

// Пока не понимаю, какое должно быть хранилище
// создаю структуру UrlStorage, которая реализует
// интерфейс Storage. Если БД использовать нельзя,
// создам что-то другое, реализующее тот же интерфейс
package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	selectByURLIDSQL     = "SELECT url FROM Url WHERE url_id = ? AND user_token = ?"
	selectByUserTokenSQL = "SELECT url, url_id FROM Url WHERE user_token = ?"
	insertSQL            = "INSERT OR REPLACE INTO Url(url, url_id, user_token) VALUES (?, ?, ?)"
)

type SQLiteStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewSQLiteStorage(conf SQLiteConfig) (*SQLiteStorage, error) {
	InitDB(conf.DBPath, conf.ClearOnStart)
	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", conf.DBPath, err)
	}
	return &SQLiteStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *SQLiteStorage) AddURL(url, urlID, userToken string) error {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(url, urlID, userToken)
	if err != nil {
		return err
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *SQLiteStorage) GetURLByID(urlID, userToken string) (storage.Record, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByURLIDSQL, urlID, userToken)
	if err != nil {
		return storage.Record{}, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	// Забираем url из первой строки
	rec := storage.Record{URLID: urlID, UserToken: userToken}
	if err := rows.Scan(&rec.URL); err != nil {
		return storage.Record{}, err
	}
	if err := rows.Err(); err != nil {
		return storage.Record{}, err
	}
	return rec, nil
}

func (s *SQLiteStorage) GetURLsByUser(userToken string) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.db.Query(selectByUserTokenSQL, userToken)
	if err != nil {
		return nil, err
	}
	// не забудем закрыть объект!
	defer rows.Close()
	for rows.Next() {
		rec := storage.Record{UserToken: userToken}
		if err := rows.Scan(&rec.URL, &rec.URLID); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if len(results) == 0 {
		return nil, storage.ErrURLWasNotFound
	}
	return results, nil
}

// Закрывает соединение с SQLite
func (s *SQLiteStorage) Close() {
	s.db.Close()
}
