package util_test

import (
	"bytes"
	"crypto/md5"
	"testing"

	"github.com/nextsurfer/word/internal/pkg/util"
)

func md5Sum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func TestAES16CBC(t *testing.T) {
	plaintext := "this is a plain text"
	key := "this is a key"
	var err error
	plaintextbytes := []byte(plaintext)
	plaintextbytes, err = util.AES16CBCEncrypt(plaintextbytes, []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	plainbytes, err := util.AES16CBCDecrypt(plaintextbytes, []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(plaintext), plainbytes) {
		t.Fatal("decrypted data not equal to origin text")
	}
}
