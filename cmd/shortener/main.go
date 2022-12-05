package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("local.env")
	s.RunServer()
}
