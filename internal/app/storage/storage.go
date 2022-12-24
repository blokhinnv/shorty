// Пакет с интерфейсом хранилища данных
package storage

import (
	"errors"
)

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(url, urlID string, userID uint32) error
	// Получает URL по ID+userID
	GetURLByID(urlID string) (Record, error)
	// Получает URLs по токену пользователя
	GetURLsByUser(userID uint32) ([]Record, error)
	// Закрывает хранилище
	Close()
}

var ErrURLWasNotFound = errors.New("requested URL was not found")
