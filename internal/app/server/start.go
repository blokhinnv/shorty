// Package server contains the logic for creating, configuring, and starting the server.
package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme/autocert"

	"github.com/blokhinnv/shorty/internal/app/log"
)

// Creates a http.Server object ready to support HTTPS.
func prepareHTTPS(r chi.Router, serverAddress string) *http.Server {
	manager := &autocert.Manager{
		// directory to store certificates
		Cache: autocert.DirCache("cache-dir"),
		// a function that accepts the Terms of Service of the certificate publisher
		Prompt: autocert.AcceptTOS,
		// list of domains for which certificates will be supported
		// TODO: can't test it because I don't have domain?
		// https://community.letsencrypt.org/t/can-i-test-lets-encrypt-client-on-localhost/15627
		HostPolicy: autocert.HostWhitelist("shorty.ru", "www.shorty.ru"),
	}
	server := &http.Server{
		Addr:      serverAddress,
		Handler:   r,
		TLSConfig: manager.TLSConfig(),
	}
	return server
}

// RunServer creates the store and starts the server.
func RunServer(cfg *config.ServerConfig, ctx context.Context) {
	s, err := database.NewDBStorage(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer s.Close(ctx)
	routerCloseCh := make(chan struct{}, 1)
	r := routes.NewRouter(s, cfg, routerCloseCh)
	log.Printf("Starting server with config %+v\n", cfg)

	var server *http.Server
	if cfg.EnableHTTPS {
		server = prepareHTTPS(r, cfg.ServerAddress)
	} else {
		server = &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: r,
		}
	}

	go func() {
		if cfg.EnableHTTPS {
			server.ListenAndServeTLS("", "")
		} else {
			server.ListenAndServe()
		}
	}()
	<-ctx.Done()
	// signal to finish deleting goroutines
	routerCloseCh <- struct{}{}
	log.Println("Shutting down server gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("shutdown error: %v\n", err)
		os.Exit(1)
	}
	// ensure goroutines are finished
	<-routerCloseCh
	log.Println("Bye!")
}
