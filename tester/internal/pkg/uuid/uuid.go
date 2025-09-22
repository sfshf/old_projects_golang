package uuid

import (
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

func NewUUIDHexEncoding() string {
	uuid := uuid.New()
	var buf [32]byte
	hex.Encode(buf[:], uuid[:])
	return strings.ToUpper(string(buf[:]))
}
