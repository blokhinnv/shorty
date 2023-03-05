package main

// не забыть стартанут сервер redis
// docker run --name redis-test-instance -p 6379:6379 -d redis

import (
	"flag"

	log "github.com/sirupsen/logrus"

	s "github.com/blokhinnv/shorty/internal/app/server"
	"github.com/blokhinnv/shorty/internal/app/server/config"
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
	flagCfg := config.FlagConfig{}
	parseConfig(&flagCfg)
	serverCfg, err := config.NewServerConfig(&flagCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.RunServer(serverCfg)
}
