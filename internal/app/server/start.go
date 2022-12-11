// Пакет для создания и настройки сервера
package server

import (
	"log"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"
)

// Создает хранилище и запускает сервер
func RunServer() {
	s := database.NewDBStorage()
	defer s.Close()
	r := routes.NewRouter(s)
	cfg := config.GetServerConfig()
	log.Printf("Starting server with config %+v\n", cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
