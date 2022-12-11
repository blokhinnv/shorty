package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"os"

	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/joho/godotenv"
)

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
	s.RunServer()
}
