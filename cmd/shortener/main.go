package main

import (
	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("local.env")
	s.RunServer()
}
