package util

import (
	"crypto/rand"
	"math/big"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateID() generates a random 8 character long base62 encoded string.
func GenerateID() (string, error) {
	id := make([]byte, 8)
	for i := range id {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}

		id[i] = alphabet[n.Int64()]
	}

	return string(id), nil
}
