// Пакет с интерфейсом хранилища данных
package storage

import "errors"

// Интерфейс для хранилища
type Storage interface {
	// добавляет URL в хранилище
	AddUrl(url string) int64
	// Получает URL по ID
	GetUrlById(id int64) (string, error)
	// Возвращает ID URL по его строковому представлению
	GetIdByUrl(url string) (int64, error)
}

var (
	ErrUrlWasNotFound = errors.New("requested URL was not found")
	ErrIdWasNotFound  = errors.New("requested ID was not found")
)
