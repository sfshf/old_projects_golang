package util_test

import (
	"log"
	"testing"

	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestGetRandomInt(t *testing.T) {
	ri, err := util.GetRandomInt(0, 100000)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(ri)
}

func TestGenerateNickname(t *testing.T) {
	rn, err := util.GenerateNickname("", 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	log.Println(rn)
}
