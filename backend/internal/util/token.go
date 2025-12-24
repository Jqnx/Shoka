package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
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

// TODO: JWK caching
func GetUserFromRequest(c *gin.Context) uuid.UUID {
	keyset, err := jwk.Fetch(c.Request.Context(), "http://localhost:3000/api/auth/jwks")
	if err != nil {
		return uuid.Nil
	}

	token, err := jwt.ParseRequest(c.Request, jwt.WithKeySet(keyset))
	if err != nil {
		return uuid.Nil
	}

	currentTime := time.Now()
	expTime, exists := token.Expiration()
	if !exists {
		return uuid.Nil
	}

	if currentTime.After(expTime) {
		return uuid.Nil
	}

	uid, exists := token.Subject()
	if !exists {
		return uuid.Nil
	}

	id, err := uuid.Parse(uid)
	if err != nil {
		return uuid.Nil
	}
	return id
}
