// Пакет для создания БД - хранилища URL

// Пока не понимаю, какое должно быть хранилище
// создаю структуру UrlStorage, которая реализует
// интерфейс Storage. Если БД использовать нельзя,
// создам что-то другое, реализующее тот же интерфейс
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	selectByURLIDSQL  = "SELECT url, user_id FROM Url WHERE url_id = ?"
	selectByUserIDSQL = "SELECT url, url_id FROM Url WHERE user_id = ?"
	insertSQL         = "INSERT OR REPLACE INTO Url(url, url_id, user_id) VALUES (?, ?, ?)"
)

type SQLiteStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewSQLiteStorage(conf SQLiteConfig) (*SQLiteStorage, error) {
	InitDB(conf.DBPath, conf.ClearOnStart)
	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", conf.DBPath, err)
	}
	return &SQLiteStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *SQLiteStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	stmt, err := s.db.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, url, urlID, userID)
	if err != nil {
		return err
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *SQLiteStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	// Получаем строки
	rows, err := s.db.QueryContext(ctx, selectByURLIDSQL, urlID)
	if err != nil {
		return storage.Record{}, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Next подготовит результат и вернет True, если строки есть
	if !rows.Next() {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	// Забираем url из первой строки
	rec := storage.Record{URLID: urlID}
	if err := rows.Scan(&rec.URL, &rec.UserID); err != nil {
		return storage.Record{}, err
	}
	if err := rows.Err(); err != nil {
		return storage.Record{}, err
	}
	return rec, nil
}

// Получает URLs по ID пользователя
func (s *SQLiteStorage) GetURLsByUser(
	ctx context.Context,
	userID uint32,
) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.db.QueryContext(ctx, selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// не забудем закрыть объект!
	defer rows.Close()
	for rows.Next() {
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	if len(results) == 0 {
		return nil, storage.ErrURLWasNotFound
	}
	return results, nil
}

// Добавляет пакет URLов в хранилище
func (s *SQLiteStorage) AddURLBatch(
	ctx context.Context,
	urlIDs map[string]string,
	userID uint32,
) error {
	// шаг 1 — объявляем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()
	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	for url, urlID := range urlIDs {
		// шаг 3 — указываем, что каждая запись будет добавлена в транзакцию
		if _, err := stmt.ExecContext(ctx, url, urlID, userID); err != nil {
			log.Println("unable to add row: ", err)
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}
	// шаг 4 — сохраняем изменения
	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
	}
	return nil
}

// Закрывает соединение с SQLite
func (s *SQLiteStorage) Close(ctx context.Context) {
	s.db.Close()
}

// Проверяет соединение с хранилищем
func (s *SQLiteStorage) Ping(ctx context.Context) bool {
	return s.db.Ping() == nil
}
