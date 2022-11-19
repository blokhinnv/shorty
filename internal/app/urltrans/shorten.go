package urltrans

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyz-_0123456789" // алфавит в 38-й СС
	base    = 38
)

// Проверяет, является ли строка URL
func isURL(s string) bool {
	_, err := url.ParseRequestURI(string(s))
	return err == nil
}

// Переводит число в 38-ую СС
func toShortenBase(urlID int64) string {
	var shortURL strings.Builder
	for urlID > 0 {
		shortURL.WriteByte(letters[urlID%base])
		urlID = urlID / base
	}
	return shortURL.String()
}

// Возвращает укороченный URL
func GetShortURL(s storage.Storage, url string) (string, error) {
	// Если не URL, то укорачивать не будет
	if !isURL(url) {
		return "", fmt.Errorf("%v is not an URL", string(url))
	}

	urlID, err := s.GetIDByURL(url)
	// Если в базе такой URL есть, то берем его ID
	// Если нет - добавляем строчку в БД
	if err == storage.ErrIDWasNotFound {
		log.Printf("Creating new row for url=%s\n", url)
		urlID = s.AddURL(url)
	} else if err != nil {
		return "", err
	}
	// Сокращаем
	shortURL := toShortenBase(urlID)

	return shortURL, nil
}
