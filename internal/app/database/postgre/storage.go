// Пакет для создания БД - хранилища URL
package database

import (
	"context"
	"log"
	"os"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/jackc/pgx/v5"
)

const (
	selectByURLIDSQL  = "SELECT url, user_id FROM Url WHERE url_id = $1"
	selectByUserIDSQL = "SELECT url, url_id FROM Url WHERE user_id = $1"
	insertSQL         = "INSERT INTO Url(url, url_id, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, url_id) DO NOTHING;"
)

type PostgreStorage struct {
	conn *pgx.Conn
}

// Конструктор нового хранилища URL
func NewPostgreStorage(conf PostgreConfig) (*PostgreStorage, error) {
	conn, err := pgx.Connect(context.Background(), conf.DatabaseDSN)
	if err != nil {
		log.Fatalf("can't access to DB %s: %v\n", conf.DatabaseDSN, err)
		os.Exit(1)
	}
	InitDB(conn, conf.ClearOnStart)

	return &PostgreStorage{conn}, nil
}

// Метод для добавления нового URL в БД
func (s *PostgreStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	_, err := s.conn.Exec(ctx, insertSQL, url, urlID, userID)
	if err != nil {
		log.Fatalf("Error while adding URL: %v", err)
		return err
	}
	log.Printf("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *PostgreStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	// Получаем строки
	err := s.conn.QueryRow(ctx, selectByURLIDSQL, urlID).
		Scan(&rec.URL, &rec.UserID)
	// любая ошибка здесь (в т.ч. ErrNoRows) означает, что результат не найден
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	return rec, nil
}

// Получает URLs по ID пользователя
func (s *PostgreStorage) GetURLsByUser(
	ctx context.Context,
	userID uint32,
) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.conn.Query(ctx, selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Проходим по всем записям методом rows.Next() до тех пор,
	// пока не пройдём все доступные результаты
	for rows.Next() {
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	// После цикла проверяем записи на потенциальные ошибки (разрыв
	// сетевого соединения с сервером базы данных в процессе получения результатов запроса)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, storage.ErrURLWasNotFound
	}
	return results, nil
}

// Добавляет пакет URLов в хранилище
func (s *PostgreStorage) AddURLBatch(
	ctx context.Context,
	urlIDs map[string]string,
	userID uint32,
) error {
	// шаг 1 — объявляем транзакцию
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback(ctx)
	// https://github.com/jackc/pgx/issues/791
	// pgx automatically prepares and caches statements by default.
	// So unless you have a very specific and unusual use case you
	// should not explicitly prepare statements.
	for url, urlID := range urlIDs {
		// шаг 3 — указываем, что каждая запись будет добавлена в транзакцию
		if _, err := tx.Exec(ctx, insertSQL, url, urlID, userID); err != nil {
			log.Println("unable to add row: ", err)
			if err = tx.Rollback(ctx); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}
	// шаг 4 — сохраняем изменения
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
	}
	return nil
}

// Закрывает соединение с Postgre
func (s *PostgreStorage) Close(ctx context.Context) {
	s.conn.Close(ctx)
}

// Проверяет соединение с хранилищем
func (s *PostgreStorage) Ping(ctx context.Context) bool {
	return s.conn.Ping(ctx) == nil
}
