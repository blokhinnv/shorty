// Package server contains the logic for creating, configuring, and starting the server.
package server

import (
	"context"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// RunServer creates the store and starts the server.
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
