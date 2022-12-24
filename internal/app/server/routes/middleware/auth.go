package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
)

type Auth struct {
	secretKey []byte
}

const UserTokenCookieName = "UserToken"
const UserTokenCtxKey = ContextStringKey(UserTokenCookieName)
const nBytesForID = 4

// генерирует случайную последовательность байт
func (m *Auth) generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Устанавливает cookie на основе подписи
func (m *Auth) setCookie(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	userToken, err := m.generateToken()
	if err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		Name:  UserTokenCookieName,
		Value: userToken,
	}
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)
	log.Printf("Set new cookie %s=%s", cookie.Name, cookie.Value)
	return &cookie, nil
}

// Генерирует подпись для данных
func (m *Auth) generateHMAC(data []byte) []byte {
	h := hmac.New(sha256.New, m.secretKey)
	h.Write(data)
	return h.Sum(nil)
}

// Генерирует подпись
func (m *Auth) generateToken() (string, error) {
	// 4 байта - ID пользователя (данные)
	id, err := m.generateRandom(nBytesForID)
	if err != nil {
		return "", err
	}
	// Подпись для ID пользователя
	sign := m.generateHMAC(id)
	// Токен: 4 байта ID + подпись для ID
	token := append(id, sign...)
	return hex.EncodeToString(token), nil
}

// Проверяет куки (токен)
func (m *Auth) verifyCookie(r *http.Request, cookie *http.Cookie) (bool, error) {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return false, nil
	}
	// первые 4 байта - ID пользователя
	// Получаем для них подпись с секретным ключом сервера
	sign := m.generateHMAC(data[:nBytesForID])
	// Сверяем то, что пришло в cookie, с настоящей подписью
	return hmac.Equal(sign, data[nBytesForID:]), nil
}

// Обработчик middleware
func (m *Auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserTokenCookieName)
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				cookie, _ = m.setCookie(w, r)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
		} else {
			if verified, _ := m.verifyCookie(r, cookie); verified {
				log.Printf("Authentification is successful")
			} else {
				log.Printf("Authentification is not successful")
				cookie, _ = m.setCookie(w, r)
			}
		}
		ctx := context.WithValue(
			r.Context(),
			ContextStringKey(UserTokenCookieName),
			cookie.Value,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Конструктор middleware
func NewAuth(key []byte) *Auth {
	return &Auth{secretKey: key}
}
