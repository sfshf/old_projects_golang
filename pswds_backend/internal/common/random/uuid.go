package random

import (
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

// NewUUIDString use google/uuid to create a uuid string
func NewUUIDString() string {
	uuid := uuid.New()
	return strings.ToUpper(uuid.String())
}

// NewUUIDHexEncoding use google/uuid to create a uuid, return a raw hex encoding type
func NewUUIDHexEncoding() string {
	uuid := uuid.New()
	var buf [32]byte
	hex.Encode(buf[:], uuid[:])
	return strings.ToUpper(string(buf[:]))
}
