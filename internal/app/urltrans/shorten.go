package urltrans

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/cespare/xxhash/v2"
)

const (
	letters = "0123456789abcdefghijklmnopqrstuvwxyz-_" // алфавит в 38-й СС
	base    = 38
)

// Проверяет, является ли строка URL
func isURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// Переводит число в 38-ую СС
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

// Возвращает укороченный URL
func GetShortURL(s storage.Storage, url, host string) (string, error) {
	// Если не URL, то укорачивать не будет
	if !isURL(url) {
		return "", fmt.Errorf("not an URL: %s ", url)
	}
	urlID := toShortenBase(xxhash.Sum64String(url))
	s.AddURL(url, urlID)
	// Сокращаем
	shortURL := fmt.Sprintf("http://%v/%v", host, urlID)
	return shortURL, nil
}
