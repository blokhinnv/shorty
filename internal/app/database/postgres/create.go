package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL-запрос для создания таблицы для Url
const createSQL = `
DROP TABLE IF EXISTS Url;
CREATE TABLE Url(
	encoding_id SERIAL PRIMARY KEY,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id BIGINT NOT NULL,
	added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE
);
CREATE UNIQUE INDEX idx_url ON Url(url);
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
func InitDB(conn *pgxpool.Pool, clearOnStart bool) {
	var exists bool
	if err := conn.QueryRow(context.Background(), existsSQL).Scan(&exists); err != nil {
		log.Fatalf("can't create table Url: %v\n", err)
	}
	if !exists || clearOnStart {
		if _, err := conn.Exec(context.Background(), createSQL); err != nil {
			log.Fatalf("can't create table Url: %v\n", err)
		}
	}
}
