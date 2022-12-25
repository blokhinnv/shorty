package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

// SQL-запрос для создания таблицы для Url
const createSQL = `
DROP TABLE IF EXISTS Url;
CREATE TABLE Url(
	encoding_id SERIAL PRIMARY KEY,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id BIGINT NOT NULL,
	added TIMESTAMP,
	requested_at TIMESTAMP
);
CREATE UNIQUE INDEX idx_url ON Url(url, url_id);
`
const existsSQL = `
SELECT EXISTS (
    SELECT FROM
        pg_tables
    WHERE
        schemaname = 'public' AND
        tablename  = 'Url'
);
`

// При инициализации создадим БД, если ее не существует
func InitDB(conn *pgx.Conn, clearOnStart bool) {
	var exists bool
	if err := conn.QueryRow(context.Background(), existsSQL).Scan(&exists); err != nil {
		log.Fatalf("can't create table Url: %v\n", err)
		os.Exit(1)
	}
	if !exists || clearOnStart {
		if _, err := conn.Exec(context.Background(), createSQL); err != nil {
			log.Fatalf("can't create table Url: %v\n", err)
			os.Exit(1)
		}
	}
}
