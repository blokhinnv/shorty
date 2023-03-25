// Package middleware contains middleware implementations for server operations.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/blokhinnv/shorty/internal/app/log"
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

// generateUserID generates a random user ID.
func (m *Auth) generateUserID(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// setCookie sets a cookie based on the signature.
func (m *Auth) setCookie(w http.ResponseWriter, r *http.Request) *http.Cookie {
	userToken, err := m.generateToken()
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

// generateHMAC generates a signature for the data.
func (m *Auth) generateHMAC(data []byte) []byte {
	h := hmac.New(sha256.New, m.secretKey)
	h.Write(data)
	return h.Sum(nil)
}

// generateToken generates a token.
func (m *Auth) generateToken() (string, error) {
	// 4 bytes - user ID (data)
	id, err := m.generateUserID(nBytesForID)
	if err != nil {
		return "", err
	}
	// Signature for user ID
	sign := m.generateHMAC(id)
	// Token: 4 bytes ID signature for ID
	token := append(id, sign...)
	return hex.EncodeToString(token), nil
}

// verifyCookie verifies cookies (token).
func (m *Auth) verifyCookie(w http.ResponseWriter, cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	// first 4 bytes - user ID
	// Get a signature for them with the server's secret key
	sign := m.generateHMAC(data[:nBytesForID])
	// Check what came in the cookie against the real signature
	return hmac.Equal(sign, data[nBytesForID:])
}

// extractID extracts the ID from the cookie.
func (m *Auth) extractID(w http.ResponseWriter, cookie *http.Cookie) uint32 {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return 0
	}
	id := binary.BigEndian.Uint32(data[:nBytesForID])
	return id
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
