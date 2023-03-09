// Пакет middleware содержит реализации middleware для работы сервера
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

// Константы для работы middleware
const (
	UserTokenCookieName = "UserToken"
	UserIDCtxKey        = ContextStringKey("UserID")
	nBytesForID         = 4
)

// Auth - структура для middleware авторизации.
type Auth struct {
	secretKey []byte
}

// NewAuth - конструктор Auth middleware.
func NewAuth(key []byte) *Auth {
	return &Auth{secretKey: key}
}

// generateUserID генерирует случайный ID пользователя.
func (m *Auth) generateUserID(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// setCookie устанавливает cookie на основе подписи.
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

// generateHMAC генерирует подпись для данных.
func (m *Auth) generateHMAC(data []byte) []byte {
	h := hmac.New(sha256.New, m.secretKey)
	h.Write(data)
	return h.Sum(nil)
}

// generateToken генерирует токен.
func (m *Auth) generateToken() (string, error) {
	// 4 байта - ID пользователя (данные)
	id, err := m.generateUserID(nBytesForID)
	if err != nil {
		return "", err
	}
	// Подпись для ID пользователя
	sign := m.generateHMAC(id)
	// Токен: 4 байта ID + подпись для ID
	token := append(id, sign...)
	return hex.EncodeToString(token), nil
}

// verifyCookie проверяет куки (токен).
func (m *Auth) verifyCookie(w http.ResponseWriter, cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return false
	}
	// первые 4 байта - ID пользователя
	// Получаем для них подпись с секретным ключом сервера
	sign := m.generateHMAC(data[:nBytesForID])
	// Сверяем то, что пришло в cookie, с настоящей подписью
	return hmac.Equal(sign, data[nBytesForID:])
}

// extractID извлекает ID из куки.
func (m *Auth) extractID(w http.ResponseWriter, cookie *http.Cookie) uint32 {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return 0
	}
	id := binary.BigEndian.Uint32(data[:nBytesForID])
	return id
}

// Handler возвращает обработчик middleware.
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
