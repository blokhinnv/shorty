package middleware

import (
	"compress/gzip"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// RequestGZipDecompress - middleware for compressing the response.
func RequestGZipDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			log.Printf("Decoding request ...")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			// not sure if this will work. How would you check this?
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
