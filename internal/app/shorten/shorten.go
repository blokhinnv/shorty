// Пакет shorten содержит логику сокращения URL.
package shorten

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/cespare/xxhash/v2"
)

// Алфавит СС.
const (
	letters = "0123456789abcdefghijklmnopqrstuvwxyz_" // алфавит в 38-й СС
	base    = 37
)

// isURL проверяет, является ли строка URL.
func isURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// toShortenBase переводит число в 38-ую СС.
func toShortenBase(urlUUID uint64) string {
	var shortURL strings.Builder
	if urlUUID == 0 {
		return string(letters[0])
	}
	for urlUUID > 0 {
		shortURL.WriteByte(letters[urlUUID%base])
		urlUUID = urlUUID / base
	}
	return shortURL.String()
}

// GetShortURL возвращает укороченный URL.
func GetShortURL(
	s storage.Storage,
	url string,
	userID uint32,
	baseURL string,
) (string, string, error) {
	// Если не URL, то укорачивать не будет
	if !isURL(url) {
		return "", "", fmt.Errorf("not an URL: %s ", url)
	}
	// Сокращаем
	shortURLID := toShortenBase(xxhash.Sum64String(url))
	// Генерим URL
	shortURL := fmt.Sprintf("%v/%v", baseURL, shortURLID)
	return shortURLID, shortURL, nil
}
