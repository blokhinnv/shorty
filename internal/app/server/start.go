// Пакет для создания и настройки сервера
package server

import (
	"log"
	"net/http"
	"os"

	db "github.com/blokhinnv/shorty/internal/app/database"
)

func RunServer() {
	storage, err := db.NewUrlStorage()
	if err != nil {
		log.Fatal("Can't connect to the URL storage")
		os.Exit(1)
	}
	http.Handle("/", &RootHandler{storage})
	// Сервер должен быть доступен по адресу: http://localhost:8080
	http.ListenAndServe(":8080", nil)
}
