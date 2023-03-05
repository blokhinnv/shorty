package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// SQL-запрос для создания таблицы для Url.
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

// SQL-запрос для проверки наличия таблицы для Url.
const existsSQL = `
SELECT EXISTS (
    SELECT FROM
        pg_tables
    WHERE
        schemaname = 'public' AND
        tablename  = 'Url'
);
`

// InitDB инициализирует структуру БД для дальнейшей работы.
func InitDB(conn *pgxpool.Pool, clearOnStart bool) error {
	var exists bool
	// при инициализации создадим БД, если ее не существует
	if err := conn.QueryRow(context.Background(), existsSQL).Scan(&exists); err != nil {
		return fmt.Errorf("can't create table Url: %v", err)
	}
	if !exists || clearOnStart {
		log.Infoln("INIT DB", exists, clearOnStart)
		if _, err := conn.Exec(context.Background(), createSQL); err != nil {
			return fmt.Errorf("can't create table Url: %v", err)
		}
	}
	return nil
}
