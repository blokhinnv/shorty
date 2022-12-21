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
func RunServer(flagCfg config.FlagConfig) {
	cfg := config.GetServerConfig(flagCfg)
	s := database.NewDBStorage(flagCfg)
	defer s.Close()
	r := routes.NewRouter(s, cfg)
	log.Printf("Starting server with config %+v\n", cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
