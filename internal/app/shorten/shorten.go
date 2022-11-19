package shorten

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/database"
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
func GetShortURL(url string) (string, error) {
	// Если не URL, то укорачивать не будет
	if !isUrl(url) {
		return "", fmt.Errorf("%v is not an URL", string(url))
	}

	db, err := database.NewConnection()
	if err != nil {
		return "", err
	}
	has_result, urlId, err := database.GetIdByUrl(db, url)
	if err != nil {
		return "", err
	}
	// Если в базе такой URL есть, то берем его ID
	// Если нет - добавляем строчку в БД
	if !has_result {
		log.Printf("Creating new row for url=%s\n", url)
		urlId = database.AddUrl(db, url)
	}
	// Сокращаем
	shortUrl := toShortenBase(urlId)

	return shortUrl, nil
}
