// Package storage contains a description of the data storage interface.
package storage

import (
	"context"
	"errors"
)

// Storage errors.
var (
	ErrURLWasNotFound  = errors.New("requested URL was not found")
	ErrUniqueViolation = errors.New("duplicate key value violates unique constraint")
	ErrURLWasDeleted   = errors.New("requested url was deleted")
)

// Storage - interface for storage.
type Storage interface {
	// AddURL adds a URL to the store.
	AddURL(ctx context.Context, url, urlID string, userID uint32) error
	// AddURLBatch adds a batch of URLs to the store.
	AddURLBatch(ctx context.Context, urlIDs map[string]string, userID uint32) error
	// GetURLByID gets URL by ID.
	GetURLByID(ctx context.Context, urlID string) (Record, error)
	// GetURLsByUser gets URLs by user ID.
	GetURLsByUser(ctx context.Context, userID uint32) ([]Record, error)
	// DeleteMany flags the URL to be deleted.
	DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error
	// Ping checks the connection to the repository.
	Ping(ctx context.Context) bool
	// Clear clears the storage.
	Clear(ctx context.Context) error
	// Close closes the store.
	Close(ctx context.Context)
	// Returns DB stats.
	GetStats(ctx context.Context) (int, int, error)
}
