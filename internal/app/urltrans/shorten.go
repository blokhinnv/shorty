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
func isUrl(s string) bool {
	_, err := url.ParseRequestURI(string(s))
	return err == nil
}

// Переводит число в 38-ую СС
func toShortenBase(urlId int64) string {
	var shortUrl strings.Builder
	for urlId > 0 {
		shortUrl.WriteByte(letters[urlId%base])
		urlId = urlId / base
	}
	return shortUrl.String()
}

// Возвращает укороченный URL
func GetShortURL(s storage.Storage, url string) (string, error) {
	// Если не URL, то укорачивать не будет
	if !isUrl(url) {
		return "", fmt.Errorf("%v is not an URL", string(url))
	}

	urlId, err := s.GetIdByUrl(url)
	// Если в базе такой URL есть, то берем его ID
	// Если нет - добавляем строчку в БД
	if err == storage.ErrIdWasNotFound {
		log.Printf("Creating new row for url=%s\n", url)
		urlId = s.AddUrl(url)
	} else if err != nil {
		return "", err
	}
	// Сокращаем
	shortUrl := toShortenBase(urlId)

	return shortUrl, nil
}
