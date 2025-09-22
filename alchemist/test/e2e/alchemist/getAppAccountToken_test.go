package alchemist_test

import (
	"log"
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestGetAppAccountToken(t *testing.T) {
	// insert one app account token
	mockAppAccountToken := SlarkUser{
		AppAccountToken: util.NewUUIDString(),
		UserID:          _testAccount.ID,
	}
	if err := _alchemistGormDB.Create(&mockAppAccountToken).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _alchemistGormDB.Delete(&SlarkUser{ID: mockAppAccountToken.ID}).Error; err != nil {
			log.Println(err)
		}
	}()
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			AppAccountToken string `json:"appAccountToken"`
		} `json:"data"`
	}{}
	// send request -- get app account token
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getAppAccountToken/v1", nil, _testCookie, &respData, nil)
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
	// check data
	if respData.Data.AppAccountToken != mockAppAccountToken.AppAccountToken {
		t.Error("not prospective response data")
		return
	}
}

func TestGetAppAccountToken_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			AppAccountToken string `json:"appAccountToken"`
		} `json:"data"`
	}{}
	// send request -- get app account token
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getAppAccountToken/v1", nil, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}
