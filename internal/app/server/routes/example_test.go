package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	database "github.com/blokhinnv/shorty/internal/app/database/mock"
	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/golang/mock/gomock"
)

func Example() {
	// Setup storage ...
	t := new(testing.T)
	ctrl := gomock.NewController(t)
	s := database.NewMockStorage(ctrl)
	s.EXPECT().AddURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	// Setup request ...
	handler := GetShortURLHandlerFunc(s)
	rr := httptest.NewRecorder()
	body := bytes.NewBuffer([]byte("https://practicum.yandex.ru/learn/"))
	req, _ := http.NewRequest(http.MethodPost, "/", body)
	// Setup context ...
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.BaseURLCtxKey, "http://localhost:8080")
	ctx = context.WithValue(ctx, middleware.UserIDCtxKey, uint32(1))

	// Run
	handler(rr, req.WithContext(ctx))
	fmt.Println(rr.Body.String())

	// Output:
	// http://localhost:8080/rb1t0eupmn2_
}
