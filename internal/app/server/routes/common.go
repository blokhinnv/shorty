package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

func shortenURLLogic(
	ctx context.Context,
	w http.ResponseWriter,
	s storage.Storage,
	longURL string,
) (string, int, error) {
	baseURL, ok := ctx.Value(middleware.BaseURLCtxKey).(string)
	if !ok {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return "", http.StatusInternalServerError, fmt.Errorf(
			http.StatusText(http.StatusInternalServerError),
		)
	}
	// В этом месте уже обязательно должно быть ясно
	// для кого мы готовим ответ
	userID, ok := ctx.Value(middleware.UserIDCtxKey).(uint32)
	if !ok {
		http.Error(
			w,
			"no user id provided",
			http.StatusInternalServerError,
		)
		return "", http.StatusInternalServerError, fmt.Errorf("no user id provided")
	}

	shortURLID, shortenURL, err := shorten.GetShortURL(s, longURL, userID, baseURL)
	if err != nil {
		return "", http.StatusBadRequest, err
	}
	err = s.AddURL(ctx, longURL, shortURLID, userID)
	status := http.StatusCreated
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			status = http.StatusConflict
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return "", http.StatusBadRequest, err
		}
	}
	return shortenURL, status, nil
}
