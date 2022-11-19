package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Конструктор нового соединения с БД
func NewConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", dbFile, err)
	}
	return db, nil
}
