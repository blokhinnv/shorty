// Пакет для создания БД - хранилища URL
package postgres

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	selectByURLIDSQL      = "SELECT url, user_id, is_deleted FROM Url WHERE url_id = $1;"
	selectByUserIDSQL     = "SELECT url, url_id, is_deleted FROM Url WHERE user_id = $1;"
	insertSQL             = "INSERT INTO Url(url, url_id, user_id) VALUES ($1, $2, $3);"
	restoreSQL            = "UPDATE Url SET is_deleted=FALSE, user_id=$2 WHERE url_id=$1 AND is_deleted=TRUE;"
	deleteBatchByURLIDSQL = "UPDATE Url SET is_deleted=TRUE WHERE url_id=ANY($1) AND user_id=$2 RETURNING url;"
	uniqueViolationCode   = "23505"
	clearSQL              = "DELETE FROM Url;"
)

type PostgresStorage struct {
	conn *pgxpool.Pool
}

// Конструктор нового хранилища URL
func NewPostgresStorage(conf *PostgresConfig) (*PostgresStorage, error) {
	// до реализации удаления работало так:
	// conn, err := pgx.Connect(context.Background(), conf.DatabaseDSN)
	// if err != nil {
	// 	log.Fatalf("can't access to DB %s: %v\n", conf.DatabaseDSN, err)
	// }
	poolConfig, err := pgxpool.ParseConfig(conf.DatabaseDSN)
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL:", err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}

	InitDB(conn, conf.ClearOnStart)

	return &PostgresStorage{conn}, nil
}

// Метод для добавления нового URL в БД
func (s *PostgresStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	res, err := s.conn.Exec(ctx, restoreSQL, urlID, userID)
	if err != nil {
		log.Infof("Error while updating URL: %v", err)
		return err
	}
	n := res.RowsAffected()
	// нашли строку для восстановления
	if n > 0 {
		return nil
	}
	// нашли строку для восстановления => надо добавить
	_, err = s.conn.Exec(ctx, insertSQL, url, urlID, userID)
	if err != nil {
		log.Infof("Error while adding URL: %v", err)
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == uniqueViolationCode {
				return fmt.Errorf(
					"%w: url=%v, urlID=%v, userID=%v",
					storage.ErrUniqueViolation,
					url,
					urlID,
					userID,
				)
			}
		}
		return err
	}
	log.Infof("Added %v=>%v to storage\n", url, urlID)
	return nil
}

// Возвращает URL по его ID в БД
func (s *PostgresStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	// Получаем строки
	var isDeleted bool
	err := s.conn.QueryRow(ctx, selectByURLIDSQL, urlID).
		Scan(&rec.URL, &rec.UserID, &isDeleted)

	// любая ошибка здесь (в т.ч. ErrNoRows) означает, что результат не найден
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	if isDeleted {
		return storage.Record{}, storage.ErrURLWasDeleted
	}
	return rec, nil
}

// Получает URLs по ID пользователя
func (s *PostgresStorage) GetURLsByUser(
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
		var isDeleted bool
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID, &isDeleted); err != nil {
			return nil, err
		}
		if !isDeleted {
			results = append(results, rec)
		}
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
func (s *PostgresStorage) AddURLBatch(
	ctx context.Context,
	urlIDs map[string]string,
	userID uint32,
) error {
	batch := &pgx.Batch{}
	for url, urlID := range urlIDs {
		// pgx automatically prepares and caches statements by default
		res, err := s.conn.Exec(ctx, restoreSQL, urlID, userID)
		if err != nil {
			return err
		}
		n := res.RowsAffected()
		if n > 0 {
			continue
		}
		batch.Queue(insertSQL, url, urlID, userID)
		log.Println("added row: ", url, urlID, userID)
	}
	br := s.conn.SendBatch(ctx, batch)
	defer br.Close()
	_, err := br.Exec()
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		if pgerr.Code == uniqueViolationCode {
			return storage.ErrUniqueViolation
		}
	} else {
		return err
	}
	return nil
}

// Устанавливает отметку об удалении URL
func (s *PostgresStorage) DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error {
	rows, err := s.conn.Query(ctx, deleteBatchByURLIDSQL, urlIDs, userID)
	if err != nil {
		log.Infof("Error while deleting URL: %v", err)
		return err
	}
	// не забудем закрыть объект!
	defer rows.Close()

	// Проходим по всем записям методом rows.Next() до тех пор,
	// пока не пройдём все доступные результаты
	updated := make([]string, 0)
	var updatedURL string
	for rows.Next() {
		if err := rows.Scan(&updatedURL); err != nil {
			return err
		}
		updated = append(updated, updatedURL)
	}
	// После цикла проверяем записи на потенциальные ошибки (разрыв
	// сетевого соединения с сервером базы данных в процессе получения результатов запроса)
	if err := rows.Err(); err != nil {
		return err
	}
	log.Infof("Set %v as deleted\n", updated)
	return nil

}

// Закрывает соединение с Postgres
func (s *PostgresStorage) Close(ctx context.Context) {
	s.conn.Close()
}

// Проверяет соединение с хранилищем
func (s *PostgresStorage) Ping(ctx context.Context) bool {
	return s.conn.Ping(ctx) == nil
}

// Очищает хранилище
func (s *PostgresStorage) Clear(ctx context.Context) error {
	_, err := s.conn.Exec(ctx, clearSQL)
	return err
}
