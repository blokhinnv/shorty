// Пакет с интерфейсом хранилища данных
package storage

import (
	"errors"
)

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(url, urlID, userID string) error
	// Получает URL по ID+userID
	GetURLByID(urlID, userID string) (Record, error)
	// Получает URLs по ID пользователя
	GetURLsByUser(userID string) ([]Record, error)
	// Закрывает хранилище
	Close()
}

var ErrURLWasNotFound = errors.New("requested URL was not found")
