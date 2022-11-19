package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const (
	selectByIDSQL  = "SELECT url FROM Url WHERE url_id = ?"
	selectByUrlSQL = "SELECT url_id FROM Url WHERE url = ?"
)

// Возвращает URL по его ID в БД
func GetUrlById(db *sql.DB, id int64) (bool, string, error) {
	// Получаем строки
	rows, err := db.Query(selectByIDSQL, id)
	if err != nil {
		return false, "", err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next вернет True, если строки есть
	has_result := rows.Next()
	if !has_result {
		return false, "", nil
	}
	// Забираем url из первой строки
	var url string
	if err := rows.Scan(&url); err != nil {
		return false, "", err
	}
	return true, url, nil
}

// Возвращает ID URL по его строковому представлению
func GetIdByUrl(db *sql.DB, url string) (bool, int64, error) {
	// Получаем строки
	rows, err := db.Query(selectByUrlSQL, url)
	if err != nil {
		return false, 0, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next вернет True, если строки есть
	has_result := rows.Next()
	if !has_result {
		return false, 0, nil
	}
	// Забираем id из первой строки
	var id int64
	if err := rows.Scan(&id); err != nil {
		return false, 0, err
	}
	return true, id, nil
}
