package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// SQL-запрос для создания таблицы для Url
const createSQL = `
DROP TABLE IF EXISTS Url;
CREATE TABLE Url(
	encoding_id INTEGER PRIMARY KEY AUTOINCREMENT,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id VARCHAR NOT NULL,
	added VARCHAR,
	requested_at VARCHAR
);
CREATE UNIQUE INDEX idx_url ON Url(url, url_id);
`

// При инициализации создадим БД, если ее не существует
func InitDB(dbFile string) {
	// Проверка существования БД
	if _, err := os.Stat(dbFile); err == nil {
		return
	}
	// Создание таблицы в БД
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("can't access to DB %s: %v\n", dbFile, err)
		os.Exit(1)
	}
	if _, err = db.Exec(createSQL); err != nil {
		log.Fatalf("can't create table Url: %v\n", err)
		os.Exit(1)
	}
}
