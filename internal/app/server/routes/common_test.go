package routes

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/log"

	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// Данные пользователя для тестирования.
const (
	userID    uint32 = 2781908098
	userToken string = "a5d08c82f3815eefe7f496d9652d8a041031e6a7f89d6bb2c90e1dfc335826e5a22255c8"
)

// Настройки для блокирования перенаправления.
var (
	errRedirectBlocked = errors.New("HTTP redirect blocked")
	NoRedirectPolicy   = resty.RedirectPolicyFunc(
		func(req *http.Request, via []*http.Request) error {
			return errRedirectBlocked
		},
	)
)

// Конфиг для запуска тестов.
type TestConfig struct {
	serverCfg *config.ServerConfig
	host      string
	port      string
	baseURL   string
}

// unsetTestEnv сбрасывает все переменные окружения.
func unsetTestEnv() {
	for _, envName := range []string{
		"SQLITE_DB_PATH",
		"SQLITE_CLEAR_ON_START",
		"DATABASE_DSN",
		"PG_CLEAR_ON_START",
		"FILE_STORAGE_PATH",
		"FILE_STORAGE_CLEAR_ON_START",
		"FILE_STORAGE_TTL_ON_DISK",
		"FILE_STORAGE_TTL_IN_MEMORY",
		"SERVER_ADDRESS",
		"BASE_URL",
		"SECRET_KEY",
	} {
		os.Unsetenv(envName)
	}

}

// NewTestConfig - конструктор конфига для запуска тестов.
func NewTestConfig(envPath string) TestConfig {
	unsetTestEnv()
	godotenv.Load(envPath)
	flagCfg := config.FlagConfig{} // будем считать, что в тестах флаги не используются
	serverCfg, err := config.NewServerConfig(&flagCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	return TestConfig{
		serverCfg: serverCfg,
		host:      serverCfg.ServerAddress,
		port:      strings.Split(serverCfg.ServerAddress, ":")[1],
		baseURL:   serverCfg.BaseURL,
	}

}

// NewServerWithPort - конструктор нового сервера.
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

// IPToLocalhost - вспомогательная функция, которая заменяет 127.0.0.1 на localhost.
func IPToLocalhost(addr string) string {
	return strings.Replace(addr, "127.0.0.1", "localhost", -1)
}

// errReader - тип для теста ошибки при io.ReadAll
type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
