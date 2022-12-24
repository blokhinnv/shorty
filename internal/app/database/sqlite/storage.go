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
	selectByURLIDSQL  = "SELECT url FROM Url WHERE url_id = ? AND user_id = ?"
	selectByUserIDSQL = "SELECT url, url_id FROM Url WHERE user_id = ?"
	insertSQL         = "INSERT OR REPLACE INTO Url(url, url_id, user_id) VALUES (?, ?, ?)"
)

type SQLiteStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewSQLiteStorage(conf SQLiteConfig) (*SQLiteStorage, error) {
	InitDB(conf.DBPath)
	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", conf.DBPath, err)
	}
	return &SQLiteStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *SQLiteStorage) AddURL(url, urlID, userID string) error {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(url, urlID, userID)
	if err != nil {
		return err
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *SQLiteStorage) GetURLByID(urlID, userID string) (storage.Record, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByURLIDSQL, urlID, userID)
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
	rec := storage.Record{URLID: urlID, UserID: userID}
	if err := rows.Scan(&rec.URL); err != nil {
		return storage.Record{}, err
	}
	if err := rows.Err(); err != nil {
		return storage.Record{}, err
	}
	return rec, nil
}

func (s *SQLiteStorage) GetURLsByUser(userID string) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.db.Query(selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// не забудем закрыть объект!
	defer rows.Close()
	for rows.Next() {
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, nil
}

// Закрывает соединение с SQLite
func (s *SQLiteStorage) Close() {
	s.db.Close()
}
