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

func TestUnregister(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()
	testPoints := 10

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
	mockReferralPoint := ReferralPoint{
		UserID: _testAccount.ID,
		App:    testAppID,
		Points: int32(testPoints),
	}
	if err := _alchemistGormDB.Create(&mockReferralPoint).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralPoint{ID: mockReferralPoint.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/unregister/v1", nil, _testCookie, &respData, nil)
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
	var referralCode ReferralCode
	if err := _alchemistGormDB.Table(TableNameReferralCode).
		Where(`user_id = ? AND app = ? AND referral_code = ? AND deleted_at > 0`,
			_testAccount.ID, testAppID, testReferralCode).
		First(&referralCode).Error; err != nil {
		t.Error(err)
		return
	}
	var referralPoint ReferralPoint
	if err := _alchemistGormDB.Table(TableNameReferralPoint).
		Where(`user_id = ? AND app = ? AND deleted_at > 0`,
			_testAccount.ID, testAppID).
		First(&referralPoint).Error; err != nil {
		t.Error(err)
		return
	}
	if referralCode.ID != mockReferralCode.ID || referralPoint.ID != mockReferralPoint.ID {
		t.Error("not prospective response data")
		return
	}
}

func TestUnregister_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/unregister/v1", nil, nil, &respData, nil)
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
