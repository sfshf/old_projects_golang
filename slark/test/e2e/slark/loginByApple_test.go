package slark_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestLoginByApple(t *testing.T) {
	ctx := context.Background()
	// test data
	testNickname := "TestLoginByApple"
	testEmail := testNickname + "@TestLoginByApple.net"
	testUserIdentifier := "TestLoginByApple"

	reqData := struct {
		Email          string `json:"email"`
		UserIdentifier string `json:"userIdentifier"`
	}{
		Email:          testEmail,
		UserIdentifier: testUserIdentifier,
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
	resp, err := postJsonRequest(_kongDNS+"/slark/user/loginByApple/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.UserID <= 0 || respData.Data.Email != testEmail {
		t.Error("not prospective response data")
		return
	}
	// check data
	var user SlkUser
	if err := _gormDB.Table(TableNameSlkUser).
		Where("nickname = ? AND email = ? AND deleted_at = 0",
			testNickname, testEmail).
		First(&user).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&user).Error; err != nil {
			t.Error(err)
		}
	}()
	var thirdParty SlkThirdParty
	if err := _gormDB.Table(TableNameSlkThirdParty).
		Where("user_id = ? AND deleted_at = 0", user.ID).
		First(&thirdParty).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&thirdParty).Error; err != nil {
			t.Error(err)
		}
	}()
	var session SlkSession
	if err := _gormDB.Table(TableNameSlkSession).
		Where("user_id = ? AND device_id = ? AND deleted_at = 0",
			user.ID, "NOID").
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
