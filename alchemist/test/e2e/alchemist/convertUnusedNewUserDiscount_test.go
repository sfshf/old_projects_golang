package alchemist_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestConvertUnusedNewUserDiscount(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.NewUUIDString()
	testTimestamp := time.Now()
	testBilledTimes := 11
	testReferralPoints := 100

	newUserDiscountState := NewUserDiscountState{
		App:            testAppID,
		UserID:         _testAccount.ID,
		ReferralCode:   testReferralCode,
		StartDate:      testTimestamp.UnixMilli(),
		BilledTimes:    int32(testBilledTimes),
		RemainingTimes: 12 - int32(testBilledTimes),
	}
	if err := _alchemistGormDB.Create(&newUserDiscountState).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&NewUserDiscountState{ID: newUserDiscountState.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	mockReferralPoint := ReferralPoint{
		App:    testAppID,
		UserID: _testAccount.ID,
		Points: int32(testReferralPoints),
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

	reqData := struct {
		AppID string `json:"appID"`
	}{
		AppID: testAppID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/convertUnusedNewUserDiscount/v1", &reqData, _testCookie, &respData, nil)
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
	var referralLog ReferralLog
	if err := _alchemistGormDB.Table(TableNameReferralLog).
		Where(`app = ? AND referral_point_id = ? AND type = ? AND reason = ? AND deleted_at = 0`,
			testAppID, mockReferralPoint.ID, consts.ReferralLogTypeGain, consts.ReferralLogReasonConvertPoints).
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
}

func TestConvertUnusedNewUserDiscount_EmptySession(t *testing.T) {
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
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/convertUnusedNewUserDiscount/v1", &reqData, nil, &respData, nil)
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
