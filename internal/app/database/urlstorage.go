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
	selectByUrlSQL = "SELECT url_id FROM Url WHERE url = ?"
	insertSQL      = "INSERT INTO Url(url) VALUES (?)"
)

type UrlStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewUrlStorage() (*UrlStorage, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", dbFile, err)
	}
	return &UrlStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *UrlStorage) AddUrl(url string) int64 {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		panic("can't prepare insert query\n")
	}
	res, err := stmt.Exec(url)
	if err != nil {
		panic("can't execute insert query\n")
	}
	urlId, _ := res.LastInsertId()
	return urlId
}

// Возвращает URL по его ID в БД
func (s *UrlStorage) GetUrlById(id int64) (string, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByIDSQL, id)
	if err != nil {
		return "", err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return "", storage.ErrUrlWasNotFound
	}
	// Забираем url из первой строки
	var url string
	if err := rows.Scan(&url); err != nil {
		return "", err
	}
	return url, nil
}

// Возвращает ID URL по его строковому представлению
func (s *UrlStorage) GetIdByUrl(url string) (int64, error) {
	// Получаем строки
	rows, err := s.db.Query(selectByUrlSQL, url)
	if err != nil {
		return -1, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return -1, storage.ErrIdWasNotFound
	}
	// Забираем id из первой строки
	var id int64
	if err := rows.Scan(&id); err != nil {
		return -1, err
	}
	return id, nil
}
