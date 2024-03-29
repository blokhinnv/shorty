// Пакет main - точка входа для сервера
//
// Пример запуска под windows: FOR /F %i IN ('git rev-parse HEAD') DO SET BUILD_COMMIT=%i && SET BUILD_DATE=%DATE% %TIME% && go run -ldflags "-X main.buildVersion=v0.0.1 -X 'main.buildDate=%BUILD_DATE%' -X 'main.buildCommit=%BUILD_COMMIT%'" .
package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/joho/godotenv"

	"github.com/blokhinnv/shorty/internal/app/server/config"
	grpc "github.com/blokhinnv/shorty/internal/app/server/grpc"
	"github.com/blokhinnv/shorty/internal/app/server/http"
)

// Глобальные переменные сборки
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// parseFlags - парсит флаги в структуру config.FlagConfig
func parseFlags(cfg *config.FlagConfig) {
	flag.StringVar(&cfg.ServerAddress, "a", "", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "base url")
	flag.StringVar(
		&cfg.FileStoragePath,
		"f",
		"",
		"file where the data is stored",
	)
	flag.StringVar(&cfg.SecretKey, "k", "", "secret key to sign uid cookies")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "postgres connect string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "whether to enable HTTPS or not")

	flag.StringVar(&cfg.JSONConfigPath, "c", "", "path to json config (shorthand)")
	flag.StringVar(&cfg.JSONConfigPath, "config", "", "path to json config")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "CIDR")
	flag.BoolVar(&cfg.StartGRPC, "g", false, "Wheither to start GRPC or HTTP server")
	flag.Parse()
}

// printBuildInfo - выводит сообщение о версии, дате и коммите билда при старте.
func printBuildInfo() {
	coalesce := func(args ...string) string {
		for _, a := range args {
			if a != "" {
				return a
			}
		}
		return ""
	}
	buildVersion = coalesce(buildVersion, "N/A")
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		coalesce(buildVersion, "N/A"),
		coalesce(buildDate, "N/A"),
		coalesce(buildCommit, "N/A"),
	)
}

func main() {
	printBuildInfo()
	// для разработки использую air, его приходится запускать их корня
	// проекта, т.к. мне нужно мониторить изменения всех файлов проекта
	// настроил air так, что собранный сервер запускается с аргументом dev,
	// так что тут я могу понять, какие переменные брать
	// если стартовать без air из папки /shortener, то все норм
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		godotenv.Load("dev.env")
	} else {
		godotenv.Load("local.env")
	}
	// флаги надо собрать в одном месте на старте
	// и прокидывать через кучу слоев....
	// раньше это была глобальная переменная для пакета

	flagCfg := config.FlagConfig{}
	parseFlags(&flagCfg)
	serverCfg, err := config.NewServerConfig(&flagCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()
	if flagCfg.StartGRPC {
		grpc.RunGRPCServer(shutdownCtx, serverCfg)
	} else {
		http.RunHTTPServer(shutdownCtx, serverCfg)
	}
}
