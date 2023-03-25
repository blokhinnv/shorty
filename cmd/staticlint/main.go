package main

import (
	"io"
	"os"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/pkg/shortycheck"
	"github.com/pelletier/go-toml/v2"
)

// stchkConfig - конфиг для staticcheck
type stchkConfig struct {
	Checks []string
}

// main - точка входа приложения.
func main() {
	// читаем конфиг
	cfgFile, err := os.Open("staticcheck.conf")
	if err != nil {
		// если конфиг не найден, запускаем без анализаторов staticcheck
		log.Infoln("unable to open config file")
		shortycheck.RunMultichecker([]string{})
	} else {
		// получаем анализаторы staticcheck
		cfgBytes, err := io.ReadAll(cfgFile)
		if err != nil {
			log.Fatal("unable to read config file")
		}
		var cfg stchkConfig
		err = toml.Unmarshal(cfgBytes, &cfg)
		if err != nil {
			log.Fatal("unable to unmarshal config file")
		}
		// запускаем с доп. анализаторами
		shortycheck.RunMultichecker(cfg.Checks)
	}

}
