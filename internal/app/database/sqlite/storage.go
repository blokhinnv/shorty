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
	selectByIDSQL    = "SELECT url FROM Url WHERE url_id = ?"
	selectByURLSQL   = "SELECT url_id FROM Url WHERE url = ?"
	insertSQL        = "INSERT OR REPLACE INTO Url(url, url_id) VALUES (?, ?)"
	maxEncodingIDSQL = "SELECT COALESCE(MAX(encoding_id), 0) FROM Url "
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
func (s *SQLiteStorage) AddURL(url, urlID string) {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		panic("can't prepare insert query\n")
	}
	_, err = stmt.Exec(url, urlID)
	if err != nil {
		panic("can't execute insert query\n")
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
}

// Возвращает URL по его ID в БД
func (s *SQLiteStorage) GetURLByID(urlID string) (string, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByIDSQL, urlID)
	if err != nil {
		return "", err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return "", storage.ErrURLWasNotFound
	}
	// Забираем url из первой строки
	var url string
	if err := rows.Scan(&url); err != nil {
		return "", err
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return url, nil
}

// Закрывает соединение с SQLite
func (s *SQLiteStorage) Close() {
	s.db.Close()
}
