// Пакет для создания БД - хранилища URL
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

const (
	selectByURLIDSQL  = "SELECT url, user_id, is_deleted FROM Url WHERE url_id = ?"
	selectByUserIDSQL = "SELECT url, url_id, is_deleted FROM Url WHERE user_id = ?"
	insertSQL         = "INSERT INTO Url(url, url_id, user_id) VALUES (?, ?, ?)"
	deleteByURLIDSQL  = "UPDATE Url SET is_deleted=TRUE WHERE url_id=? AND user_id=? RETURNING url;"
	clearSQL          = "DELETE FROM Url"
	restoreSQL        = "UPDATE Url SET is_deleted=FALSE, user_id=? WHERE url_id=? AND is_deleted=TRUE;"
)

type SQLiteStorage struct {
	db *sql.DB
}

// Конструктор нового хранилища URL
func NewSQLiteStorage(conf *SQLiteConfig) (*SQLiteStorage, error) {
	err := InitDB(conf.DBPath, conf.ClearOnStart)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", conf.DBPath, err)
	}
	return &SQLiteStorage{db}, nil
}

// Метод для добавления нового URL в БД
func (s *SQLiteStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	res, err := s.db.ExecContext(ctx, restoreSQL, userID, urlID)
	if err != nil {
		log.Infof("Error while updating URL: %v", err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		log.Infof("Error while updating URL: %v", err)
		return err
	}
	// восстановили запись => добавлять не надо
	if n > 0 {
		return nil
	}
	// надо добавить
	_, err = s.db.ExecContext(ctx, insertSQL, url, urlID, userID)
	if err != nil {
		log.Infof("Error while adding URL: %v", err)
		if sqlerr, ok := err.(sqlite3.Error); ok {
			if sqlerr.Code == sqlite3.ErrConstraint {
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
func (s *SQLiteStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	var isDeleted bool
	err := s.db.QueryRowContext(ctx, selectByURLIDSQL, urlID).
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
		var isDeleted bool
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID, &isDeleted); err != nil {
			return nil, err
		}
		// После цикла проверяем записи на потенциальные ошибки (разрыв
		// сетевого соединения с сервером базы данных в процессе получения результатов запроса)
		if !isDeleted {
			results = append(results, rec)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
	var violationErr error
	stmtInsert, err := s.db.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()
	stmtRestore, err := s.db.PrepareContext(ctx, restoreSQL)
	if err != nil {
		return err
	}
	defer stmtRestore.Close()

	for url, urlID := range urlIDs {
		// пытаемся сбросить флаг об удалении
		res, err := stmtRestore.ExecContext(ctx, userID, urlID)
		if err != nil {
			log.Println("unable to update row: ", err)
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			log.Println("unable to get rows affected: ", err)
			return err
		}
		// нашли удаленную запись для восстановления => добавлять не надо
		if n > 0 {
			continue
		}
		// не нашли запись для восстановления
		if _, err := stmtInsert.ExecContext(ctx, url, urlID, userID); err != nil {
			log.Println("unable to add row: ", err)
			var sqlerr sqlite3.Error
			// она могла быть, но не удаленная, тогда будет нарушение индекса
			if errors.As(err, &sqlerr) {
				if sqlerr.Code == sqlite3.ErrConstraint {
					violationErr = fmt.Errorf(
						"%w: url=%v, urlID=%v, userID=%v",
						storage.ErrUniqueViolation,
						url,
						urlID,
						userID,
					)
				}
			} else {
				return err
			}
		}
		log.Println("added row: ", url, urlID, userID)
	}
	if violationErr != nil {
		return violationErr
	}
	return nil
}

func (s *SQLiteStorage) DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, deleteByURLIDSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	updated := make([]string, 0)
	for _, urlID := range urlIDs {
		err = func() error {
			var deletedURL string
			rows, err := stmt.QueryContext(ctx, urlID, userID)
			if err != nil {
				if errRollback := tx.Rollback(); errRollback != nil {
					log.Fatalf("update drivers: unable to rollback: %v", errRollback)
				}
				return err
			}
			defer rows.Close()
			if !rows.Next() {
				return nil
			}
			if err := rows.Scan(&deletedURL); err != nil {
				return err
			}
			if err := rows.Err(); err != nil {
				return err
			}
			updated = append(updated, deletedURL)
			return nil
		}()
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
	}
	log.Infof("Set %v as deleted\n", updated)
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

// Очищает хранилище
func (s *SQLiteStorage) Clear(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, clearSQL)
	return err
}
