package uuid_test

import (
	"log"
	"testing"

	"github.com/nextsurfer/tester/internal/pkg/uuid"
)

func TestUuid(t *testing.T) {
	log.Println(uuid.NewUUIDHexEncoding())
}
