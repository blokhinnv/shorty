// Пакет, в котором реализованы методы для сокращения и разворачивания URL
package urltrans

import (
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Возвращает запрашиваемый по ID URL, если он существует
func GetOriginalURL(s storage.Storage, url_id string) (string, error) {
	url, err := s.GetURLByID(url_id)
	if err != nil {
		return "", err
	}
	return url, nil
}
