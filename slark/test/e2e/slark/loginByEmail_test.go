package slark_test

import (
	"encoding/hex"
	"log"
	"net/http"
	"testing"

	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
)

func TestLoginByEmailUsingFaultPassword(t *testing.T) {
	// ctx := context.Background()
	// test data
	testEmail := "gavin@n1xt.net"
	testPassword := "qwer1234"
	h := sha3.New256()
	if _, err := h.Write([]byte(testPassword)); err != nil {
		t.Error(err)
		return
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))
	// mock data
	newSlkUser := SlkUser{
		Email:        testEmail,
		Nickname:     "TestLoginByEmailUsingFaultPassword",
		PasswordHash: testPasswordHash,
	}
	if err := _gormDB.Create(&newSlkUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _gormDB.Delete(&SlkUser{ID: newSlkUser.ID}).Error; err != nil {
			log.Println(err)
		}
	}()

	reqData := struct {
		Email        string `json:"email"`
		PasswordHash string `json:"passwordHash"`
	}{
		Email:        testEmail,
		PasswordHash: testPasswordHash + "_fault",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/user/loginByEmail/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeInvalidPassword {
		t.Error("not prospective response data code")
		return
	}
}
