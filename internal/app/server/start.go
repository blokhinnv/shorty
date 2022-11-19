package server

import "net/http"

func RunServer() {
	http.HandleFunc("/", ShortenHandler)
	// Сервер должен быть доступен по адресу: http://localhost:8080
	http.ListenAndServe(":8080", nil)
}
