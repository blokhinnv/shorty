// Package sqlite implements SQLite-based storage.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/mattn/go-sqlite3"
)

// SQL queries to implement the necessary logic.
const (
	selectByURLIDSQL  = "SELECT url, user_id, is_deleted FROM Url WHERE url_id = ?"
	selectByUserIDSQL = "SELECT url, url_id, is_deleted FROM Url WHERE user_id = ?"
	insertSQL         = "INSERT INTO Url(url, url_id, user_id) VALUES (?, ?, ?)"
	deleteByURLIDSQL  = "UPDATE Url SET is_deleted=TRUE WHERE url_id=? AND user_id=? RETURNING url;"
	clearSQL          = "DELETE FROM Url"
	restoreSQL        = "UPDATE Url SET is_deleted=FALSE, user_id=? WHERE url_id=? AND is_deleted=TRUE;"
)

// SQLiteStorage implements the Storage interface based on SQLite.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage - A constructor for a new URL storage.
func NewSQLiteStorage(conf *SQLiteConfig) (*SQLiteStorage, error) {
	err := initDB(conf.DBPath, conf.ClearOnStart)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("can't access to DB %s: %v", conf.DBPath, err)
	}
	return &SQLiteStorage{db}, nil
}

// AddURL - method for adding a new URL to the database.
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
	// restored record => no need to add
	if n > 0 {
		return nil
	}
	// must be added
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

// GetURLByID returns a URL by its ID in the database.
func (s *SQLiteStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	var isDeleted bool
	err := s.db.QueryRowContext(ctx, selectByURLIDSQL, urlID).
		Scan(&rec.URL, &rec.UserID, &isDeleted)
	// any error here (including ErrNoRows) means no result found
	if err != nil {
		return storage.Record{}, storage.ErrURLWasNotFound
	}
	if isDeleted {
		return storage.Record{}, storage.ErrURLWasDeleted
	}
	return rec, nil
}

// GetURLsByUser gets URLs by user ID.
func (s *SQLiteStorage) GetURLsByUser(
	ctx context.Context,
	userID uint32,
) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.db.QueryContext(ctx, selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// don't forget to close the object!
	defer rows.Close()
	for rows.Next() {
		var isDeleted bool
		rec := storage.Record{UserID: userID}
		if err := rows.Scan(&rec.URL, &rec.URLID, &isDeleted); err != nil {
			return nil, err
		}
		// After the loop, check the records for potential errors (break
		// network connection to the database server in the process of getting query results)
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

// AddURLBatch adds a batch of URLs to the store.
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
		// try to reset the deletion flag
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
		// found a deleted record to restore => no need to add
		if n > 0 {
			continue
		}
		// did not find an entry to restore
		if _, err := stmtInsert.ExecContext(ctx, url, urlID, userID); err != nil {
			log.Println("unable to add row: ", err)
			var sqlerr sqlite3.Error
			// it could be, but not deleted, then there will be an index violation
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

// DeleteMany flags the URL to be deleted.
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

// Close closes the connection to SQLite.
func (s *SQLiteStorage) Close(ctx context.Context) {
	s.db.Close()
}

// Ping checks the connection to the repository.
func (s *SQLiteStorage) Ping(ctx context.Context) bool {
	return s.db.Ping() == nil
}

// Clear clears the storage.
func (s *SQLiteStorage) Clear(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, clearSQL)
	return err
}

// Returns DB stats.
func (s *SQLiteStorage) GetStats(ctx context.Context) (int, int, error) {
	var urls int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Url").
		Scan(&urls)
	if err != nil {
		return 0, 0, err
	}

	var users int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT user_id) FROM Url").
		Scan(&users)
	if err != nil {
		return 0, 0, err
	}
	return urls, users, nil
}
