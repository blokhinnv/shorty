package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

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

// Точка входа клиента
func main() {
	// адрес сервиса (как его писать, расскажем в следующем уроке)
	endpoint := "http://localhost:8080/"
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
	sendRestyRequest(endpoint, long)
}
