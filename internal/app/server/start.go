// Пакет для создания и настройки сервера
package server

import (
	"log"
	"net/http"
)

func RunServer() {
	r := NewRouter()
	// Сервер должен быть доступен по адресу: http://localhost:8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
