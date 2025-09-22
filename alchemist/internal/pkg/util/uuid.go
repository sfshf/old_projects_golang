package util

import (
	"strings"

	"github.com/google/uuid"
)

func NewUUIDString() string {
	return strings.ToLower(uuid.New().String())
}
