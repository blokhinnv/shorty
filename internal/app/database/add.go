// Пакет для взаимодействия с БД
package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const insertSQL = "INSERT INTO Url(url) VALUES (?)"

// Функция для добавления нового URL в БД
func AddUrl(db *sql.DB, url string) int64 {
	stmt, err := db.Prepare(insertSQL)
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
