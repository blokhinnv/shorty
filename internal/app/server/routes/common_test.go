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

var flagCfg = config.FlagConfig{} // будем считать, что в тестах флаги не используются
var serverCfg = config.GetServerConfig(flagCfg)
var (
	host    = serverCfg.ServerAddress
	port    = strings.Split(serverCfg.ServerAddress, ":")[1]
	baseURL = serverCfg.BaseURL
)

const (
	userID    uint32 = 159190441
	userToken string = "097d0da931dd3d38279e5f9e7703943aed867a734985d723730ddcf2f32b7047d5f67248"
)

// Конструктор нового сервера
// Нужен, чтобы убедиться, что сервер запустится на нужном нам порте
func NewServerWithPort(r chi.Router, port string) *httptest.Server {
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

// Настройки для блокирования перенаправления
var errRedirectBlocked = errors.New("HTTP redirect blocked")
var NoRedirectPolicy = resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
	return errRedirectBlocked
})

func IPToLocalhost(addr string) string {
	return strings.Replace(addr, "127.0.0.1", "localhost", -1)
}
