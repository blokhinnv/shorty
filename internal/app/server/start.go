// Пакет server содержит логику создания, настройки и запуска сервера.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	defaultLog "log"

	log "github.com/sirupsen/logrus"

	"github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/server/routes"
)

// logFormatter - кастомный формат для логгера logrus.
type logFormatter struct {
}

// Format реализует кастомный вывод сообщения.
func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(
		fmt.Sprintf(
			"%v [%v] %v\n",
			entry.Time.Format("2006/01/02 03:04:05"),
			entry.Level,
			entry.Message),
	), nil
}

// init настраивает поток и формат вывода для логгера.
func init() {
	log.SetOutput(os.Stdout)
	defaultLog.SetOutput(os.Stdout)
	log.SetFormatter(new(logFormatter))
	log.SetLevel(log.DebugLevel)
}

// RunServer создает хранилище и запускает сервер.
func RunServer(cfg *config.ServerConfig) {
	// shutdownCtx, _ := signal.NotifyContext(
	// 	context.Background(),
	// 	syscall.SIGINT,
	// 	syscall.SIGKILL,
	// )
	// go func() {
	// 	<-shutdownCtx.Done()
	// 	log.Printf("Shutting down gracefully...")
	// 	os.Exit(0)
	// }()

	s, err := database.NewDBStorage(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer s.Close(context.Background())
	r := routes.NewRouter(s, cfg)
	log.Printf("Starting server with config %+v\n", cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
