package alchemist_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestCheckReferralCodeAndDevice(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()
	testDeviceToken := `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`

	// insert one referral code record
	mockReferralCode := ReferralCode{
		ReferralCode: testReferralCode,
		UserID:       _testAccount.ID,
		App:          testAppID,
		JoinDate:     time.Now(),
	}
	if err := _alchemistGormDB.Create(&mockReferralCode).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralCode{ID: mockReferralCode.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	mockUserRegisteredOnOldDevice := UserRegisteredOnOldDevice{
		App:          testAppID,
		UserID:       _testAccount.ID,
		ReferralCode: testReferralCode,
	}
	if err := _alchemistGormDB.Create(&mockUserRegisteredOnOldDevice).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&UserRegisteredOnOldDevice{ID: mockUserRegisteredOnOldDevice.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		AppID        string `json:"appID"`
		DeviceToken  string `json:"deviceToken"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        testAppID,
		DeviceToken:  testDeviceToken,
		ReferralCode: testReferralCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			CodeValid   bool `json:"codeValid"`
			DeviceValid bool `json:"deviceValid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/checkReferralCodeAndDevice/v1", &reqData, _testCookie, &respData, nil)
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
	if !respData.Data.CodeValid {
		t.Error("not prospective response data")
		return
	}
}

// func TestCheckReferralCodeAndDevice_IncorrectToken(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"
// 	testReferralCode := util.GenerateReferralCode()
// 	testDeviceToken := `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`

// 	// insert one referral code record
// 	mockReferralCode := ReferralCode{
// 		ReferralCode: testReferralCode,
// 		UserID:       _testAccount.ID,
// 		App:          testAppID,
// 		JoinDate:     time.Now(),
// 	}
// 	if err := _alchemistGormDB.Create(&mockReferralCode).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		// clear test data
// 		if err := _alchemistGormDB.Delete(&ReferralCode{ID: mockReferralCode.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()

// 	reqData := struct {
// 		AppID        string `json:"appID"`
// 		DeviceToken  string `json:"deviceToken"`
// 		ReferralCode string `json:"referralCode"`
// 	}{
// 		AppID:        testAppID,
// 		DeviceToken:  testDeviceToken,
// 		ReferralCode: testReferralCode,
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			CodeValid   bool `json:"codeValid"`
// 			DeviceValid bool `json:"deviceValid"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/checkReferralCodeAndDevice/v1", &reqData, _testCookie, &respData, nil)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		t.Error("not prospective response code")
// 		return
// 	}
// 	if respData.Code != response.StatusCodeBadRequest {
// 		t.Error("not prospective response data code")
// 		return
// 	}
// 	// check data
// 	if respData.DebugMessage != "devicecheck: bad request: Missing or incorrectly formatted device token payload" {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestCheckReferralCodeAndDevice_EmptySession(t *testing.T) {
	// mock data
	reqData := struct {
		AppID        string `json:"appID"`
		DeviceToken  string `json:"deviceToken"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        "TestCheckReferralCodeAndDevice_EmptySession_AppID",
		DeviceToken:  "TestCheckReferralCodeAndDevice_EmptySession_DeviceToken",
		ReferralCode: "TestCheckReferralCodeAndDevice_EmptySession_ReferralCode",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/checkReferralCodeAndDevice/v1", &reqData, nil, &respData, nil)
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
