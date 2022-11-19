// Пакет для создания БД - хранилища URL

// Пока не понимаю, какое должно быть хранилище
// создаю структуру UrlStorage, которая реализует
// интерфейс Storage. Если БД использовать нельзя,
// создам что-то другое, реализующее тот же интерфейс
package database

import (
	"database/sql"
	"fmt"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	selectByIDSQL  = "SELECT url FROM Url WHERE url_id = ?"
	selectByURLSQL = "SELECT url_id FROM Url WHERE url = ?"
	insertSQL      = "INSERT INTO Url(url) VALUES (?)"
)

type URLStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewURLStorage() (*URLStorage, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", dbFile, err)
	}
	return &URLStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *URLStorage) AddURL(url string) int64 {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		panic("can't prepare insert query\n")
	}
	res, err := stmt.Exec(url)
	if err != nil {
		panic("can't execute insert query\n")
	}
	urlID, _ := res.LastInsertId()
	return urlID
}

// Возвращает URL по его ID в БД
func (s *URLStorage) GetURLByID(id int64) (string, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByIDSQL, id)
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

// Возвращает ID URL по его строковому представлению
func (s *URLStorage) GetIDByURL(url string) (int64, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByURLSQL, url)
	if err != nil {
		return -1, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return -1, storage.ErrIDWasNotFound
	}
	// Забираем id из первой строки
	var id int64
	if err := rows.Scan(&id); err != nil {
		return -1, err
	}
	if err := rows.Err(); err != nil {
		return -1, err
	}
	return id, nil
}
