// Package auth contains the logic for obtaining the user ID.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// Amount of bytes used to store ID.
const nBytesForID = 4

// generateUserID generates a random user ID.
func generateUserID(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// generateHMAC generates a signature for the data.
func generateHMAC(data []byte, secretKey []byte) []byte {
	h := hmac.New(sha256.New, secretKey)
	h.Write(data)
	return h.Sum(nil)
}

// GenerateToken generates a token.
func GenerateToken(secretKey []byte) (string, error) {
	// 4 bytes - user ID (data)
	id, err := generateUserID(nBytesForID)
	if err != nil {
		return "", err
	}
	// Signature for user ID
	sign := generateHMAC(id, secretKey)
	// Token: 4 bytes ID signature for ID
	token := append(id, sign...)
	return hex.EncodeToString(token), nil
}

// verifyToken verifies cookies (token).
func VerifyToken(token []byte, secretKey []byte) bool {
	// first 4 bytes - user ID
	// Get a signature for them with the server's secret key
	sign := generateHMAC(token[:nBytesForID], secretKey)
	// Check what came in the cookie against the real signature
	return hmac.Equal(sign, token[nBytesForID:])
}

// extractID extracts the ID from the cookie.
func ExtractID(token []byte) uint32 {
	id := binary.BigEndian.Uint32(token[:nBytesForID])
	return id
}
