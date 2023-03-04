package sqlite

import (
	"database/sql"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

// SQL-запрос для создания таблицы для Url
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

// При инициализации создадим БД, если ее не существует
func InitDB(dbFile string, clearOnStart bool) {
	// Проверка существования БД
	if _, err := os.Stat(dbFile); err == nil && clearOnStart {
		err := os.Remove(dbFile)
		if err != nil {
			log.Fatalf("ClearOnStart error: %v", err)
		}
	}
	// Создание таблицы в БД
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("can't access to DB %s: %v\n", dbFile, err)
	}
	defer db.Close()
	if _, err = db.Exec(createSQL); err != nil {
		log.Fatalf("can't create table Url: %v\n", err)
	}
}
