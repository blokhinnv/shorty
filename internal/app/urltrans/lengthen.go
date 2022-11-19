// Пакет, в котором реализованы методы для сокращения и разворачивания URL
package urltrans

import (
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Возвращает запрашиваемый по ID URL, если он существует
func GetOriginalURL(s storage.Storage, id int64) (string, error) {
	url, err := s.GetURLByID(id)
	if err != nil {
		return "", err
	}
	return url, nil
}
