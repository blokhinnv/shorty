package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"log"
	"net/http"
	"strings"
)

type Auth struct {
	signer hash.Hash
}

const UserIDCookieName = "UserID"
const UserIDCtxKey = ContextStringKey(UserIDCookieName)

// Определяет пользователя на основе IP
// пока не понимаю, норм это или нет
// но каким-то образом мне нужно идентифицировать пользователя
// для проверки подписи
func (m *Auth) getClientIdentity(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Устанавливает cookie на основе подписи
func (m *Auth) setCookie(w http.ResponseWriter, r *http.Request) *http.Cookie {
	cookieValue := hex.EncodeToString(m.generateSignature(m.getClientIdentity(r)))
	cookie := http.Cookie{
		Name:  UserIDCookieName,
		Value: cookieValue,
	}
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)
	log.Printf("Set new cookie %s=%s", cookie.Name, cookie.Value)
	return &cookie
}

// Генерирует подпись
func (m *Auth) generateSignature(remoteAddr string) []byte {
	m.signer.Reset()
	m.signer.Write([]byte(remoteAddr))
	s := m.signer.Sum(nil)
	return s
}

// Проверяет куки (подпись)
func (m *Auth) verifyCookie(r *http.Request, cookie *http.Cookie) bool {
	recomputedSignature := m.generateSignature(m.getClientIdentity(r))
	gotSignature, err := hex.DecodeString(cookie.Value)
	if err != nil {
		log.Printf("Error decoding cookie %v %v", cookie.Value, err)
		return false
	}
	return hmac.Equal(recomputedSignature, gotSignature)
}

// Обработчик middleware
func (m *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserIDCookieName)
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				cookie = m.setCookie(w, r)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
		} else {
			if m.verifyCookie(r, cookie) {
				log.Printf("Authentification is successful")
			} else {
				log.Printf("Authentification is not successful")
				cookie = m.setCookie(w, r)
			}
		}
		ctx := context.WithValue(
			r.Context(),
			ContextStringKey(UserIDCtxKey),
			cookie.Value,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Конструктор middleware
func NewAuth(key []byte) *Auth {
	return &Auth{hmac.New(sha256.New, key)}
}
