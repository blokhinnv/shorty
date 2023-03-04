package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/joho/godotenv"
)

func parseConfig(cfg *config.FlagConfig) {
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
	flag.Parse()
}

func main() {
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
	parseConfig(&flagCfg)
	serverCfg, err := config.NewServerConfig(&flagCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.RunServer(serverCfg)
}
