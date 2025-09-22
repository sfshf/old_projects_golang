package random

import (
	"github.com/google/uuid"
)

func GenerateRequestId() string {
	return uuid.New().String()
}

func GenerateUUID() string {
	return uuid.New().String()
}
