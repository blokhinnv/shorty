// Пакет для создания и настройки сервера
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"
)

type LogFormatter struct {
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(
		fmt.Sprintf(
			"%v [%v] %v\n",
			entry.Time.Format("2006/01/02 03:04:05"),
			entry.Level,
			entry.Message),
	), nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(new(LogFormatter))
}

// Создает хранилище и запускает сервер
func RunServer(cfg *config.ServerConfig) {
	s, err := database.NewDBStorage(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer s.Close(context.Background())
	r := routes.NewRouter(s, cfg)
	log.Printf("Starting server with config %+v\n", cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
