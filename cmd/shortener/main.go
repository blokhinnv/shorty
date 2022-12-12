package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"flag"
	"os"

	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/joho/godotenv"
)

var flagCfg = config.FlagConfig{}

func init() {
	flag.StringVar(&flagCfg.ServerAddress, "a", "http://localhost:8080", "server address")
	flag.StringVar(&flagCfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&flagCfg.FileStoragePath, "f", "db.jsonl", "file where the data is stored")
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
	flag.Parse()
	s.RunServer(flagCfg)
}
