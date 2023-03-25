// Package shorten contains the URL shortening logic.
package shorten

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cespare/xxhash/v2"
)

// Alphabet SS.
const (
	letters = "0123456789abcdefghijklmnopqrstuvwxyz_" // alphabet in 38th SS
	base    = 37
)

// isURL checks if the string is a URL.
func isURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// toShortenBase translates the number to the 38th CC.
func toShortenBase(urlUUID uint64) string {
	var shortURL strings.Builder
	for urlUUID > 0 {
		shortURL.WriteByte(letters[urlUUID%base])
		urlUUID = urlUUID / base
	}
	return shortURL.String()
}

// GetShortURL returns the shortened URL.
func GetShortURL(
	url string,
	userID uint32,
	baseURL string,
) (string, string, error) {
	// If not URL, then will not shorten
	if !isURL(url) {
		return "", "", fmt.Errorf("not an URL: %s ", url)
	}
	// Reduce
	shortURLID := toShortenBase(xxhash.Sum64String(url))
	// Generate URL
	shortURL := fmt.Sprintf("%v/%v", baseURL, shortURLID)
	return shortURLID, shortURL, nil
}
