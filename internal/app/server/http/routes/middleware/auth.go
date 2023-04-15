// Package middleware contains middleware implementations for server operations.
package middleware

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/auth"
)

// Constants for middleware operation
const (
	UserTokenCookieName = "UserToken"
	UserIDCtxKey        = ContextStringKey("UserID")
	nBytesForID         = 4
)

// Auth - structure for authorization middleware.
type Auth struct {
	secretKey []byte
}

// NewAuth - Auth middleware constructor.
func NewAuth(key []byte) *Auth {
	return &Auth{secretKey: key}
}

// setCookie sets a cookie based on the signature.
func (m *Auth) setCookie(w http.ResponseWriter, r *http.Request) *http.Cookie {
	userToken, err := auth.GenerateToken(m.secretKey)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	cookie := http.Cookie{
		Name:  UserTokenCookieName,
		Value: userToken,
	}
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)
	log.Printf("Set new cookie %s=%s", cookie.Name, cookie.Value)
	return &cookie
}

// verifyCookie verifies cookies (token).
func (m *Auth) verifyCookie(w http.ResponseWriter, cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	return auth.VerifyToken(data, m.secretKey)
}

// extractID extracts the ID from the cookie.
func (m *Auth) extractID(w http.ResponseWriter, cookie *http.Cookie) uint32 {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return 0
	}
	return auth.ExtractID(data)
}

// Handler returns a middleware handler.
func (m *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserTokenCookieName)

		if err == nil {
			verified := m.verifyCookie(w, cookie)
			if verified {
				log.Printf("Authentification is successful")
			} else {
				log.Printf("Authentification is not successful")
				cookie = m.setCookie(w, r)
			}
		} else if errors.Is(err, http.ErrNoCookie) {
			cookie = m.setCookie(w, r)
		}
		userID := m.extractID(w, cookie)
		log.Printf("Added userID=%v to the context", userID)
		ctx := context.WithValue(
			r.Context(),
			UserIDCtxKey,
			userID,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
