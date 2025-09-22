package slark_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestSendLoginEmailCode(t *testing.T) {
	ctx := context.Background()
	// test data
	testEmail := "e2e-TestSendLoginEmailCode@n1xt.net"
	testKey := util.PrefixLoginCode + "NOID:" + testEmail
	defer func() {
		if err := _redisCli.Del(ctx, testKey).Err(); err != nil {
			t.Error(err)
		}
	}()

	// check redis data
	ecs, err := getLoginEmailCaptchas()
	if err != nil {
		t.Error(err)
		return
	}
	for _, ec := range ecs {
		if ec.Email == testEmail {
			t.Error("test email has cached")
			return
		}
	}

	newSlkUser := SlkUser{
		Email:    testEmail,
		Nickname: "TestSendLoginEmailCode",
	}
	if err := _gormDB.Create(&newSlkUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _gormDB.Delete(&SlkUser{ID: newSlkUser.ID}).Error; err != nil {
			t.Error(err)
		}
	}()

	reqData := struct {
		Email string `json:"email"`
	}{
		Email: testEmail,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/sendLoginEmailCode/v1", &reqData, nil, &respData, nil)
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
	time.Sleep(3 * time.Second)
	// check redis data
	ecs, err = getLoginEmailCaptchas()
	if err != nil {
		t.Error(err)
		return
	}
	var has bool
	for _, ec := range ecs {
		if ec.Email == testEmail {
			has = true
			break
		}
	}
	if !has {
		t.Error("not prospective response data")
		return
	}
}
