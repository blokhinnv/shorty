package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/blokhinnv/shorty/internal/app/log"

	"github.com/blokhinnv/shorty/internal/app/server/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// DeleteURLsHandler - структура для реализации хендлера на удаление URL.
type DeleteURLsHandler struct {
	s              storage.Storage
	delURLsCh      chan Job
	expireDuration time.Duration
}

// Job - задачи на удаления, которые обрабатываются горутинами.
type Job struct {
	URL    string
	UserID uint32
}

// NewDeleteURLsHandler - конструктор DeleteURLsHandler.
func NewDeleteURLsHandler(s storage.Storage, delURLsChBufSize int) *DeleteURLsHandler {
	h := &DeleteURLsHandler{
		s:              s,
		delURLsCh:      make(chan Job, delURLsChBufSize),
		expireDuration: 100 * time.Millisecond,
	}
	h.Loop()
	return h
}

// DeleteURLs подготавливает батч для удаления и передает хранилищу.
func (h *DeleteURLsHandler) DeleteURLs(jobsToDelete []Job) {
	if len(jobsToDelete) == 0 {
		return
	}
	jobsByUser := make(map[uint32][]string)
	for _, job := range jobsToDelete {
		jobsByUser[job.UserID] = append(jobsByUser[job.UserID], job.URL)
	}
	for userID, userJobs := range jobsByUser {
		go func(userID uint32, userJobs []string) {
			err := h.s.DeleteMany(context.Background(), userID, userJobs)
			if err != nil {
				log.Printf("Error while deleting urls: %v\n", err)
			}
		}(userID, userJobs)
	}
}

// Loop - основной цикл горутин для удаления URL.
func (h *DeleteURLsHandler) Loop() {
	go func() {
		jobs := make([]Job, 0)
		ticker := time.NewTicker(h.expireDuration)
		for {
			select {
			case job, ok := <-h.delURLsCh:
				if !ok {
					return
				}
				jobs = append(jobs, job)
			case <-ticker.C:
				h.DeleteURLs(jobs)
				jobs = make([]Job, 0)
			}
		}
	}()
}

// Handler - реализация эндпоинта DELETE /api/user/urls.
// Принимает список идентификаторов сокращённых URL для
// удаления в формате: [ "a", "b", "c", "d", ...].
func (h *DeleteURLsHandler) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	// Читаем тело запроса
	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't read body: %v", err.Error()), http.StatusBadRequest)
		return
	}
	bodyDecoded := make([]string, 0)
	if err = json.Unmarshal(bodyRaw, &bodyDecoded); err != nil {
		http.Error(w, fmt.Sprintf("Can't decode body: %e", err), http.StatusBadRequest)
		return
	}
	userID, ok := ctx.Value(middleware.UserIDCtxKey).(uint32)
	if !ok {
		http.Error(
			w,
			"no user id provided",
			http.StatusInternalServerError,
		)
		return
	}

	go func() {
		for _, url := range bodyDecoded {
			h.delURLsCh <- Job{URL: url, UserID: userID}
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}
