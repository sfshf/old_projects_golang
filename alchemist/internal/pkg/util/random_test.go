package util_test

import (
	"log"
	"testing"

	"github.com/nextsurfer/alchemist/internal/pkg/util"
)

// go test -v -run ^TestGenerateReferralCode$ -count=1 ./internal/app/utils/random_test.go
func TestGenerateReferralCode(t *testing.T) {
	log.Println(util.GenerateReferralCode())
}
