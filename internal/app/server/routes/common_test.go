package routes

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

const port = "8080"

var host = fmt.Sprintf("localhost:%v", port)

// Конструктор нового сервера
// Нужен, чтобы убедиться, что сервер запустится на нужном нам порте
func NewServerWithPort(r chi.Router, port string) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
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
