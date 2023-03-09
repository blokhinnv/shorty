// Пакет storage содержит описание интерфейса хранилища данных.
package storage

import (
	"context"
	"errors"
)

// Ошибки хранилища.
var (
	ErrURLWasNotFound  = errors.New("requested URL was not found")
	ErrUniqueViolation = errors.New("duplicate key value violates unique constraint")
	ErrURLWasDeleted   = errors.New("requested url was deleted")
)

// Интерфейс для хранилища.
type Storage interface {
	// AddURL добавляет URL в хранилище.
	AddURL(ctx context.Context, url, urlID string, userID uint32) error
	// AddURLBatch добавляет пакет URLов в хранилище.
	AddURLBatch(ctx context.Context, urlIDs map[string]string, userID uint32) error
	// GetURLByID получает URL по ID.
	GetURLByID(ctx context.Context, urlID string) (Record, error)
	// GetURLsByUser получает URLs по ID пользователя.
	GetURLsByUser(ctx context.Context, userID uint32) ([]Record, error)
	// DeleteMany устанавливает отметку об удалении URL.
	DeleteMany(ctx context.Context, userID uint32, urlIDs []string) error
	// Ping проверяет соединение с хранилищем.
	Ping(ctx context.Context) bool
	// Clear очищает хранилище.
	Clear(ctx context.Context) error
	// Close закрывает хранилище.
	Close(ctx context.Context)
}
