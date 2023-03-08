package main

//lint:file-ignore U1000 игнорируем неиспользуемый код

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

var compress bool

func init() {
	flag.BoolVar(&compress, "c", false, "whether to compress request body")
}

// Вариант с "ручным" созданием клиента, запроса и т.д.
func sendHTTPRequest(endpoint, long string) {
	// контейнер данных для запроса
	data := url.Values{}
	// заполняем контейнер данными
	data.Set("url", long)
	// конструируем HTTP-клиент
	client := &http.Client{}
	// конструируем запрос
	// запрос методом POST должен, кроме заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	// в большинстве случаев отлично подходит bytes.Buffer
	request, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(long),
		// bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// в заголовках запроса сообщаем, что данные кодированы стандартной URL-схемой
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// печатаем код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// и печатаем его
	fmt.Println(string(body))
}

// Вариант с отправкой запроса при помощи resty
func sendRestyRequest(endpoint, long string) {
	log.Printf("Sending raw request...")
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Content-Length", strconv.Itoa(len(long))).
		SetBody(long).
		Post(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(resp.Body()))
}

func compressLongURL(long string) string {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		panic(err.Error())
	}
	_, err = w.Write([]byte(long))
	if err != nil {
		panic(err.Error())
	}
	err = w.Close()
	if err != nil {
		panic(err.Error())
	}
	return b.String()
}

// Вариант с отправкой запроса при помощи resty c компрессией данных
func sendRestyRequestCompessed(endpoint, long string) {
	log.Printf("Sending compressed request...")
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Content-Length", strconv.Itoa(len(long))).
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressLongURL(long)).
		Post(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(resp.Body()))
}

// Точка входа клиента
func main() {
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		godotenv.Load("dev.env")
	} else {
		godotenv.Load("local.env")
	}
	flag.Parse()
	// адрес сервиса (как его писать, расскажем в следующем уроке)
	flagCfg := config.FlagConfig{}
	serverCfg, err := config.NewServerConfig(&flagCfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	endpoint := fmt.Sprintf("http://%v/", serverCfg.ServerAddress)
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	long = strings.TrimSpace(long)
	// sendHTTPRequest(endpoint, long)
	if compress {
		sendRestyRequestCompessed(endpoint, long)
	} else {
		sendRestyRequest(endpoint, long)
	}
}
