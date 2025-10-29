package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GetAuthTokenFromHeader(header string) string {
	token := strings.Split(header, " ")[1]
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashString := hex.EncodeToString(tokenHash[:])
	return tokenHashString
}
