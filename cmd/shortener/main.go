package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"os"

	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		godotenv.Load("dev.env")
	} else {
		godotenv.Load("local.env")
	}
	s.RunServer()
}
