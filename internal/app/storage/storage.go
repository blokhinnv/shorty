// Пакет с интерфейсом хранилища данных
package storage

import (
	"errors"
)

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(url, urlID string)
	// Получает URL по ID
	GetURLByID(id string) (string, error)
	// Закрывает хранилище
	Close()
}

var ErrURLWasNotFound = errors.New("requested URL was not found")
