// Пакет server содержит логику создания, настройки и запуска сервера.
package server

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// RunServer создает хранилище и запускает сервер.
func RunServer(cfg *config.ServerConfig) {
	s, err := database.NewDBStorage(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer s.Close(context.Background())
	r := routes.NewRouter(s, cfg)
	log.Printf("Starting server with config %+v\n", cfg)
	http.ListenAndServe(cfg.ServerAddress, r)
}
