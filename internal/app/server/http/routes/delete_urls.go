package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/blokhinnv/shorty/internal/app/log"

	"github.com/blokhinnv/shorty/internal/app/server/http/routes/middleware"
	"github.com/blokhinnv/shorty/internal/app/storage"
)

// DeleteURLsHandler - a structure for implementing a URL delete handler.
type DeleteURLsHandler struct {
	s              storage.Storage
	delURLsCh      chan Job
	expireDuration time.Duration
	routerCloseCh  chan struct{}
}

// Job - deletion tasks that are processed by goroutines.
type Job struct {
	URL    string
	UserID uint32
}

// NewDeleteURLsHandler - DeleteURLsHandler constructor.
func NewDeleteURLsHandler(
	s storage.Storage,
	delURLsChBufSize int,
	routerCloseCh chan struct{},
) *DeleteURLsHandler {
	h := &DeleteURLsHandler{
		s:              s,
		delURLsCh:      make(chan Job, delURLsChBufSize),
		expireDuration: 100 * time.Millisecond,
		routerCloseCh:  routerCloseCh,
	}
	h.loop()
	return h
}

// deleteURLs prepares the batch for deletion and passes it to the repository.
func (h *DeleteURLsHandler) deleteURLs(jobsToDelete []Job) {
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

// loop - the main goroutine loop for removing URLs.
func (h *DeleteURLsHandler) loop() {
	go func() {
		jobs := make([]Job, 0)
		ticker := time.NewTicker(h.expireDuration)
	out:
		for {
			select {
			case job, ok := <-h.delURLsCh:
				if !ok {
					return
				}
				jobs = append(jobs, job)
			case <-ticker.C:
				h.deleteURLs(jobs)
				jobs = make([]Job, 0)
			case <-h.routerCloseCh:
				close(h.delURLsCh)
				break out
			}
		}
		log.Info("Finishing deleting...")
		for {
			job, ok := <-h.delURLsCh
			if !ok {
				break
			}
			jobs = append(jobs, job)
		}
		h.deleteURLs(jobs)
		h.routerCloseCh <- struct{}{}
	}()
}

// Handler - implementation of the DELETE /api/user/urls endpoint.
// Accepts a list of short URL identifiers for
// deletion in the format: [ "a", "b", "c", "d", ...].
func (h *DeleteURLsHandler) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	// Read request body
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
