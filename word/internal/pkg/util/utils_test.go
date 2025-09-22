package util_test

import (
	"testing"

	"github.com/nextsurfer/word/internal/pkg/util"
)

func TestAES16CBCEncryptAndDecrypt(t *testing.T) {
	key := "secret-key"
	src := "this is the source text"
	enc, err := util.AES16CBCEncrypt([]byte(src), []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	dst, err := util.AES16CBCDecrypt(enc, []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	if string(dst) != src {
		t.Fatal("failed to encrypt or descrypt")
	}
}
