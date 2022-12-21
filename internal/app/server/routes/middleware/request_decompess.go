package middleware

import (
	"compress/gzip"
	"log"
	"net/http"
)

// Middleware для декодирования запроса
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
			// не уверен, что это даст эффект. Как бы это проверить?
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
