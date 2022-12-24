// Пакет с интерфейсом хранилища данных
package storage

import (
	"errors"
)

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(url, urlID, userToken string) error
	// Получает URL по ID+userToken
	GetURLByID(urlID string) (Record, error)
	// Получает URLs по токену пользователя
	GetURLsByUser(userToken string) ([]Record, error)
	// Закрывает хранилище
	Close()
}

var ErrURLWasNotFound = errors.New("requested URL was not found")
