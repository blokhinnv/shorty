// Пакет с интерфейсом хранилища данных
package storage

import (
	"context"
	"errors"
)

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(ctx context.Context, url, urlID string, userID uint32) error
	// добавляет пакет URLов в хранилище
	AddURLBatch(ctx context.Context, urlIDs map[string]string, userID uint32) error
	// Получает URL по ID
	GetURLByID(ctx context.Context, urlID string) (Record, error)
	// Получает URLs по ID пользователя
	GetURLsByUser(ctx context.Context, userID uint32) ([]Record, error)
	// Проверяет соединение с хранилищем
	Ping(ctx context.Context) bool
	// Закрывает хранилище
	Close(ctx context.Context)
}

var ErrURLWasNotFound = errors.New("requested URL was not found")
