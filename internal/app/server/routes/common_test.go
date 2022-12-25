package routes

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

const (
	userID    uint32 = 2781908098
	userToken string = "a5d08c82f3815eefe7f496d9652d8a041031e6a7f89d6bb2c90e1dfc335826e5a22255c8"
)

// Настройки для блокирования перенаправления
var errRedirectBlocked = errors.New("HTTP redirect blocked")
var NoRedirectPolicy = resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
	return errRedirectBlocked
})

// Конфиг для запуска тестов
type TestConfig struct {
	serverCfg config.ServerConfig
	host      string
	port      string
	baseURL   string
}

// Конструктор конфига для запуска тестов
func NewTestConfig() TestConfig {
	var flagCfg = config.FlagConfig{} // будем считать, что в тестах флаги не используются
	var serverCfg = config.NewServerConfig(flagCfg)
	return TestConfig{
		serverCfg: serverCfg,
		host:      serverCfg.ServerAddress,
		port:      strings.Split(serverCfg.ServerAddress, ":")[1],
		baseURL:   serverCfg.BaseURL,
	}

}

// Конструктор нового сервера
// Нужен, чтобы убедиться, что сервер запустится на нужном нам порте
func NewServerWithPort(r chi.Router, host, port string) *httptest.Server {
	l, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	return ts
}

// Вспомогательная функция, которая заменяет 127.0.0.1 на localhost
func IPToLocalhost(addr string) string {
	return strings.Replace(addr, "127.0.0.1", "localhost", -1)
}
