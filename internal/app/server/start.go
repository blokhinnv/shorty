// Пакет для создания и настройки сервера
package server

import (
	"log"
	"net/http"

	db "github.com/blokhinnv/shorty/internal/app/database"
)

func RunServer() {
	dbPath := "./db.sqlite3"
	db.InitDB(dbPath)
	r := NewRouter(dbPath)
	// Сервер должен быть доступен по адресу: http://localhost:8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
