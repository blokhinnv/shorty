package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	db "github.com/blokhinnv/shorty/internal/app/database"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
	"github.com/blokhinnv/shorty/internal/app/urltrans"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Функция для заполнения хранилища примерами
func addRecords(
	t *testing.T,
	s storage.Storage,
	userID uint32,
	baseURL string,
) []ShortenedURLSAnswer {
	longURLs := []string{"https://sqliteonline.com/", "https://mail.ru/"}
	answer := make([]ShortenedURLSAnswer, len(longURLs))
	for idx, longURL := range longURLs {
		shortURL, err := urltrans.GetShortURL(s, longURL, userID, baseURL)
		require.NoError(t, err)
		answer[idx] = ShortenedURLSAnswer{URL: longURL, URLID: shortURL}
	}
	return answer
}

// Тесты для GET-запроса
func ListOfURLsTestLogic(t *testing.T, testCfg TestConfig) {
	s := db.NewDBStorage(testCfg.serverCfg)
	defer s.Close()
	r := NewRouter(s, testCfg.serverCfg)
	ts := NewServerWithPort(r, testCfg.host, testCfg.port)
	defer ts.Close()

	// Заготовка под тест: создаем хранилище, сокращаем
	// один URL, проверяем, что все прошло без ошибок
	reqURL := "http://localhost:8080/api/user/urls"
	var answer []ShortenedURLSAnswer
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			// нет данных
			name: "test_no_content",
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			// есть данные
			name: "test_ok",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		// Первую итерацию делаем по незаполненному хранилищу
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetCookie(&http.Cookie{
				Name:  middleware.UserTokenCookieName,
				Value: userToken,
			})

			res, err := client.R().Get(reqURL)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			if res.StatusCode() == http.StatusOK {
				var v []ShortenedURLSAnswer
				reader := bytes.NewReader(res.Body())
				err = json.NewDecoder(reader).Decode(&v)
				assert.NoError(t, err)
				assert.Equal(t, answer, v)
			}
		})
		// Остальные - по заполненному
		answer = addRecords(t, s, userID, testCfg.baseURL)
	}
}

func Test_ListOfURLs_SQLite(t *testing.T) {
	godotenv.Load("test_sqlite.env")
	ListOfURLsTestLogic(t, NewTestConfig())
}

func Test_ListOfURLs_Text(t *testing.T) {
	godotenv.Load("test_text.env")
	ListOfURLsTestLogic(t, NewTestConfig())
}

func Test_ListOfURLs_Postgre(t *testing.T) {
	godotenv.Load("test_postgre.env")
	ListOfURLsTestLogic(t, NewTestConfig())
}
