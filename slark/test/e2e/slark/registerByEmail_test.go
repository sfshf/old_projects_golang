package slark_test

import (
	"context"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestRegisterByEmail(t *testing.T) {
	ctx := context.Background()
	// test data
	testEmail := "TestRegisterByEmail@n1xt.net"
	testNickname := "nickname-TestRegisterByEmail"
	testKey := util.PrefixRegistrationCode + "NOID:" + testEmail
	testCaptcha := "637999"
	testPassword := "qwer1234"
	h := sha3.NewKeccak256()
	_, err := h.Write([]byte(testPassword))
	if err != nil {
		t.Error(err)
		return
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))
	// inject test data
	if err := _redisCli.Set(
		ctx,
		testKey,
		testCaptcha,
		time.Minute*5,
	).Err(); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		_ = _redisCli.Del(ctx, testKey).Err()
	}()
	// send request
	reqData := struct {
		Email        string `json:"email"`
		Nickname     string `json:"nickname"`
		PasswordHash string `json:"passwordHash"`
		Captcha      string `json:"captcha"`
	}{
		Email:        testEmail,
		Nickname:     testNickname,
		PasswordHash: testPasswordHash,
		Captcha:      testCaptcha,
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
		} `json:"data,omitempty"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/user/registerByEmail/v1", &reqData, nil, &respData, nil)
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
	// check mysql data
	var slkUser SlkUser
	if err := _gormDB.Table(TableNameSlkUser).
		Where("id = ?", respData.Data.UserID).
		Where("deleted_at = 0").
		First(&slkUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&slkUser).Error; err != nil {
			t.Error(err)
		}
	}()
	var slkSession SlkSession
	if err := _gormDB.Table(TableNameSlkSession).
		Where("user_id = ?", respData.Data.UserID).
		Where("deleted_at = 0").
		First(&slkSession).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&slkSession).Error; err != nil {
			t.Error(err)
		}
		if err := _redisCli.Del(ctx, util.PrefixSession+slkSession.SessionID).Err(); err != nil {
			t.Error(err)
		}
	}()
}
