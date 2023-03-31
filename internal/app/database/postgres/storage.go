package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/blokhinnv/shorty/internal/app/log"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries to implement the necessary logic.
const (
	selectByURLIDSQL      = "SELECT url, user_id, is_deleted FROM Url WHERE url_id = $1;"
	selectByUserIDSQL     = "SELECT url, url_id, is_deleted FROM Url WHERE user_id = $1;"
	insertSQL             = "INSERT INTO Url(url, url_id, user_id) VALUES ($1, $2, $3);"
	restoreSQL            = "UPDATE Url SET is_deleted=FALSE, user_id=$2 WHERE url_id=$1 AND is_deleted=TRUE;"
	deleteBatchByURLIDSQL = "UPDATE Url SET is_deleted=TRUE WHERE url_id=ANY($1) AND user_id=$2 RETURNING url;"
	uniqueViolationCode   = "23505"
	clearSQL              = "DELETE FROM Url;"
)

// PostgresStorage implements the Storage interface based on Postgres.
type PostgresStorage struct {
	conn *pgxpool.Pool
}

// NewPostgresStorage - A constructor for a new URL storage.
func NewPostgresStorage(conf *PostgresConfig) (*PostgresStorage, error) {
	// before the implementation of deletion, it worked like this:
	// conn, err := pgx.Connect(context.Background(), conf.DatabaseDSN)
	// if err != nil {
	// log.Fatalf("can't access DB %s: %v\n", conf.DatabaseDSN, err)
	// }
	poolConfig, err := pgxpool.ParseConfig(conf.DatabaseDSN)
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL:", err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}

	err = initDB(conn, conf.ClearOnStart)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{conn}, nil
}

// AddURL - Method for adding a new URL to the database.
func (s *PostgresStorage) AddURL(ctx context.Context, url, urlID string, userID uint32) error {
	res, err := s.conn.Exec(ctx, restoreSQL, urlID, userID)
	if err != nil {
		log.Infof("Error while updating URL: %v", err)
		return err
	}
	n := res.RowsAffected()
	// found a line to restore
	if n > 0 {
		return nil
	}
	// found a line to restore => need to add
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

// GetURLByID returns a URL by its ID in the database.
func (s *PostgresStorage) GetURLByID(ctx context.Context, urlID string) (storage.Record, error) {
	rec := storage.Record{URLID: urlID}
	// Get rows
	var isDeleted bool
	err := s.conn.QueryRow(ctx, selectByURLIDSQL, urlID).
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
func (s *PostgresStorage) GetURLsByUser(
	ctx context.Context,
	userID uint32,
) ([]storage.Record, error) {
	results := make([]storage.Record, 0)

	rows, err := s.conn.Query(ctx, selectByUserIDSQL, userID)
	if err != nil {
		return nil, err
	}
	// don't forget to close the object!
	defer rows.Close()

	// Iterate through all records with the rows.Next() method until
	// until we go through all available results
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
	// After the loop, check the records for potential errors (break
	// network connection to the database server in the process of getting query results)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, storage.ErrURLWasNotFound
	}
	return results, nil
}

// AddURLBatch adds a batch of URLs to the store.
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

// DeleteMany flags the URL to be deleted.
func (s *PostgresStorage) DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error {
	rows, err := s.conn.Query(ctx, deleteBatchByURLIDSQL, urlIDs, userID)
	if err != nil {
		log.Infof("Error while deleting URL: %v", err)
		return err
	}
	// don't forget to close the object!
	defer rows.Close()

	// Iterate through all records with the rows.Next() method until
	// until we go through all available results
	updated := make([]string, 0)
	var updatedURL string
	for rows.Next() {
		if err := rows.Scan(&updatedURL); err != nil {
			return err
		}
		updated = append(updated, updatedURL)
	}
	// After the loop, check the records for potential errors (break
	// network connection to the database server in the process of getting query results)
	if err := rows.Err(); err != nil {
		return err
	}
	log.Infof("Set %v as deleted\n", updated)
	return nil

}

// Close closes the connection to Postgres.
func (s *PostgresStorage) Close(ctx context.Context) {
	s.conn.Close()
}

// Ping checks the connection to the repository.
func (s *PostgresStorage) Ping(ctx context.Context) bool {
	return s.conn.Ping(ctx) == nil
}

// Clear clears the storage.
func (s *PostgresStorage) Clear(ctx context.Context) error {
	_, err := s.conn.Exec(ctx, clearSQL)
	return err
}
