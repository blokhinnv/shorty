package sqlite

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// createSQL - SQL-запрос для создания таблицы для URL.
const createSQL = `
DROP TABLE IF EXISTS Url;
CREATE TABLE Url(
	encoding_id INTEGER PRIMARY KEY AUTOINCREMENT,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id INT NOT NULL,
	added VARCHAR DEFAULT (datetime('now','localtime')),
	requested_at VARCHAR DEFAULT (datetime('now','localtime')),
	is_deleted BOOLEAN DEFAULT FALSE
);
CREATE UNIQUE INDEX idx_url ON Url(url);
`

// InitDB инициализирует структуру БД для дальнейшей работы
func InitDB(dbFile string, clearOnStart bool) error {
	// Проверка существования БД
	if _, err := os.Stat(dbFile); err == nil && clearOnStart {
		err := os.Remove(dbFile)
		if err != nil {
			return fmt.Errorf("clearOnStart error: %v", err)
		}
	}
	// Создание таблицы в БД
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("can't access to DB %s: %v", dbFile, err)
	}
	defer db.Close()
	if _, err = db.Exec(createSQL); err != nil {
		return fmt.Errorf("can't create table Url: %v", err)
	}
	return nil
}
