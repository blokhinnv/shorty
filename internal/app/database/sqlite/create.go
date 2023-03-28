package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// createSQL - SQL query to create a table for the URL.
const createSQL = `
CREATE TABLE IF NOT EXISTS Url(
	encoding_id INTEGER PRIMARY KEY AUTOINCREMENT,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id INT NOT NULL,
	added VARCHAR DEFAULT (datetime('now','localtime')),
	requested_at VARCHAR DEFAULT (datetime('now','localtime')),
	is_deleted BOOLEAN DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_url ON Url(url);
`

// initDB initializes the database structure for further work
func initDB(dbFile string, clearOnStart bool) error {
	// Create a table in the database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("can't access to DB %s: %v", dbFile, err)
	}
	defer db.Close()
	if _, err = db.Exec(createSQL); err != nil {
		return fmt.Errorf("can't create table Url: %v", err)
	}
	if clearOnStart {
		if _, err = db.Exec(clearSQL); err != nil {
			return fmt.Errorf("can't create table Url: %v", err)
		}
	}
	return nil
}
