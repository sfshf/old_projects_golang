package slark_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/dchest/captcha"
	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
)

func TestUpdateNickname(t *testing.T) {
	// ctx := context.Background()
	// test data
	randomDigits := captcha.RandomDigits(6)
	testNickname := fmt.Sprintf("%s%d%d%d%d%d%d",
		"TestUpdateNickname",
		randomDigits[0],
		randomDigits[1],
		randomDigits[2],
		randomDigits[3],
		randomDigits[4],
		randomDigits[5],
	)

	reqData := struct {
		Nickname string `json:"nickname"`
	}{
		Nickname: testNickname,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/updateNickname/v1", &reqData, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// NOTE: this is a global account, don't remove it.
	var user SlkUser
	if err := _gormDB.Table(TableNameSlkUser).
		Where("nickname = ? AND email = ? AND deleted_at = 0",
			testNickname, _testEmail).
		First(&user).Error; err != nil {
		t.Error(err)
		return
	}
}
