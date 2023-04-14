package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/blokhinnv/shorty/internal/app/storage"
)

type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

type GetStats struct {
	s             storage.Storage
	trustedSubnet *net.IPNet // example: 192.168.0.1 in 192.168.0.0/24
}

func NewGetStats(s storage.Storage, trustedSubnet string) *GetStats {
	if trustedSubnet == "" {
		return &GetStats{s: s, trustedSubnet: nil}
	}

	_, ipv4Net, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Fatalln(err)
	}
	return &GetStats{s: s, trustedSubnet: ipv4Net}
}

func (h *GetStats) Handler(w http.ResponseWriter, r *http.Request) {
	if h.trustedSubnet == nil {
		http.Error(
			w,
			"trusted network is not set",
			http.StatusForbidden,
		)
		return
	}
	ipStr := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)
	if ip == nil {
		http.Error(
			w,
			"failed parse ip from http header",
			http.StatusInternalServerError,
		)
		return
	}
	if !h.trustedSubnet.Contains(ip) {
		http.Error(
			w,
			fmt.Sprintf("ip %v is not in trusted network %+v", ip, h.trustedSubnet),
			http.StatusForbidden,
		)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	urls, users, err := h.s.GetStats(ctx)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	result := statsResponse{URLs: urls, Users: users}
	resultEncoded, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// .. and send with the required headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resultEncoded)
}
