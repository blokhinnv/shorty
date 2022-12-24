// Пакет, в котором реализованы методы для сокращения и разворачивания URL
package urltrans

import (
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// Возвращает запрашиваемый по ID URL, если он существует
func GetOriginalURL(s storage.Storage, urlID string) (string, error) {
	rec, err := s.GetURLByID(urlID)
	if err != nil {
		return "", err
	}
	return rec.URL, nil
}
