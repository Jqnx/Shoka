package util

import (
	"encoding/hex"

	"github.com/google/uuid"
)

func GenShortenedUUID() string {
	newUuid := uuid.New()
	buf := newUuid[0:4]
	return hex.EncodeToString(buf)
}
