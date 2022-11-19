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
	url_id INTEGER PRIMARY KEY AUTOINCREMENT,
	url TEXT NOT NULL
);
`

// SQL-запрос для сдвига первых url_id до 100000 (исключительно для демонстрационных целей)
const shiftInitialIDSQL = `INSERT INTO sqlite_sequence(seq, name) VALUES (100000, 'Url')`

// Путь к файлу БД
const dbFile = "./db.sqlite3"

// При инициализации создадим БД, если ее не существует
func init() {
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
	tx, _ := db.Begin()
	if _, err = tx.Exec(createSQL); err != nil {
		log.Fatalf("can't create table Url: %v\n", err)
		tx.Rollback()
		os.Exit(1)
	}
	if _, err = tx.Exec(shiftInitialIDSQL); err != nil {
		log.Fatalf("can't setup table Url: %v\n", err)
		tx.Rollback()
		os.Exit(1)
	}
	tx.Commit()
}
