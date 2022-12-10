// Пакет для создания и настройки сервера
package server

import (
	"log"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/routes"
)

// Создает хранилище и запускает сервер
func RunServer() {
	s := database.NewDBStorage()
	defer s.Close()
	r := routes.NewRouter(s)
	// Сервер должен быть доступен по адресу: http://localhost:8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
