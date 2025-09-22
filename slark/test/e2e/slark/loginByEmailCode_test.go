package slark_test

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestLoginByEmailCode(t *testing.T) {
	ctx := context.Background()
	// test data
	testEmail := "email@TestLoginByEmailCode.net"
	testKey := util.PrefixLoginCode + "NOID:" + testEmail
	testCode, err := util.GenerateDigitCaptchaWithStoreFuncs(ctx, func(captcha string) error {
		if err := _redisCli.Set(ctx, testKey, captcha, time.Minute*5).Err(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		_ = _redisCli.Del(ctx, testKey).Err()
	}()

	newSlkUser := SlkUser{
		Email:    testEmail,
		Nickname: "TestLoginByEmailCode",
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
		Email string `json:"email"`
		Code  string `json:"code"`
	}{
		Email: testEmail,
		Code:  testCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			UserID   int64  `json:"userID"`
			Nickname string `json:"nickname"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/loginByEmailCode/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.UserID != newSlkUser.ID || respData.Data.Email != testEmail {
		t.Error("not prospective response data")
		return
	}
	var session SlkSession
	if err := _gormDB.Table(TableNameSlkSession).
		Where("user_id = ? AND device_id = ? AND deleted_at = 0",
			respData.Data.UserID, "NOID").
		First(&session).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&session).Error; err != nil {
			t.Error(err)
		}
		if err := _redisCli.Del(ctx, util.PrefixSession+session.SessionID).Err(); err != nil {
			t.Error(err)
		}
	}()
}
