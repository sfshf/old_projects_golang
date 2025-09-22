package slark_test

import (
	"context"
	"encoding/hex"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/slark/api/response"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestUnregister(t *testing.T) {
	ctx := context.Background()
	testEmail := "Test@TestUnregister.com"
	testNickname := "TestUnregister"
	testSessionID := util.NewUUIDHexEncoding()
	testPassword := "qwer1234"
	h := sha3.NewKeccak256()
	if _, err := h.Write([]byte(testPassword)); err != nil {
		t.Error(err)
		return
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))

	// mock data
	newSlkUser := SlkUser{
		Email:        testEmail,
		PasswordHash: testPasswordHash,
		Nickname:     testNickname,
	}
	if err := _gormDB.Create(&newSlkUser).Error; err != nil {
		t.Error(err)
		return
	}
	newSlkSession := SlkSession{
		UserID:    newSlkUser.ID,
		DeviceID:  "NOID",
		SessionID: testSessionID,
	}
	if err := _gormDB.Create(&newSlkSession).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _gormDB.Delete(&newSlkUser).Error; err != nil {
			log.Println(err)
		}
		if err := _gormDB.Delete(&newSlkSession).Error; err != nil {
			log.Println(err)
		}
	}()
	if err := _redisCli.Set(ctx, util.PrefixSession+testSessionID, testSessionID, time.Second*5).Err(); err != nil {
		t.Error(err)
		return
	}
	testCookie := &http.Cookie{
		HttpOnly: false,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    testSessionID,
	}

	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/unregister/v1", nil, testCookie, &respData, nil)
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
	var user SlkUser
	if err := _gormDB.Table(TableNameSlkUser).
		Where("id = ? AND deleted_at > 0", newSlkUser.ID).
		First(&user).Error; err != nil {
		t.Error(err)
		return
	}
	var session SlkSession
	if err := _gormDB.Table(TableNameSlkSession).
		Where("user_id = ? AND session_id = ? AND deleted_at > 0",
			newSlkUser.ID, testSessionID).
		First(&session).Error; err != nil {
		t.Error(err)
		return
	}
	if err := _redisCli.Get(ctx, util.PrefixSession+testSessionID).Err(); err != redis.Nil {
		t.Error(err)
		return
	}
}
