package alchemist_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestJoinReferralProgram(t *testing.T) {
	// mock data
	testAppID := "alchemist"

	reqData := struct {
		AppID string `json:"appID"`
	}{
		AppID: testAppID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ReferralCode string `json:"referralCode"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/joinReferralProgram/v1", &reqData, _testCookie, &respData, nil)
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
		Where(`user_id = ? AND app = ? AND deleted_at = 0`,
			_testAccount.ID, testAppID).
		First(&referralCode).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralCode{ID: referralCode.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralPoint ReferralPoint
	if err := _alchemistGormDB.Table(TableNameReferralPoint).
		Where(`user_id = ? AND app = ? AND points = 10 AND deleted_at = 0`,
			_testAccount.ID, testAppID).
		First(&referralPoint).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralPoint{ID: referralPoint.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralLog ReferralLog
	if err := _alchemistGormDB.Table(TableNameReferralLog).
		Where(`user_id = ? AND app = ? AND type = ? AND reason = ? AND points = 10 AND deleted_at = 0`,
			_testAccount.ID, testAppID, consts.ReferralLogTypeGain, consts.ReferralLogReasonNewUserBilled).
		First(&referralLog).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralLog{ID: referralLog.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var freeTrialState FreeTrialState
	if err := _alchemistGormDB.Table(TableNameFreeTrialState).
		Where(`user_id = ? AND app = ? AND deleted_at = 0`,
			_testAccount.ID, testAppID).
		First(&freeTrialState).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&FreeTrialState{ID: freeTrialState.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	if respData.Data.ReferralCode != referralCode.ReferralCode {
		t.Error("not prospective response data")
		return
	}
}

func TestJoinReferralProgram_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"

	reqData := struct {
		AppID string `json:"appID"`
	}{
		AppID: testAppID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ReferralCode string `json:"referralCode"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/joinReferralProgram/v1", &reqData, nil, &respData, nil)
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
