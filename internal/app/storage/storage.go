// Пакет с интерфейсом хранилища данных
package storage

import "errors"

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddURL(url, url_id string) int64
	// Получает URL по ID
	GetURLByID(id string) (string, error)
	// Возвращает ID URL по его строковому представлению
	GetIDByURL(url string) (string, error)
	// Возвращает свободный числовой ID для кодировки URL
	GetFreeUID() (int, error)
}

var (
	ErrURLWasNotFound = errors.New("requested URL was not found")
	ErrIDWasNotFound  = errors.New("requested ID was not found")
)
