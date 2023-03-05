package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL-запрос для создания таблицы для Url.
const createSQL = `
CREATE TABLE IF NOT EXISTS Url(
	encoding_id SERIAL PRIMARY KEY,
	url VARCHAR NOT NULL,
	url_id VARCHAR NOT NULL,
	user_id BIGINT NOT NULL,
	added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_url ON Url(url);
`

// InitDB инициализирует структуру БД для дальнейшей работы.
func InitDB(conn *pgxpool.Pool, clearOnStart bool) error {
	if _, err := conn.Exec(context.Background(), createSQL); err != nil {
		return fmt.Errorf("can't create table Url: %v", err)
	}

	if clearOnStart {
		if _, err := conn.Exec(context.Background(), clearSQL); err != nil {
			return fmt.Errorf("can't clear table: %v", err)
		}
	}
	return nil
}
