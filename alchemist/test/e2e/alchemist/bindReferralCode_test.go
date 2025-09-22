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

func TestBindReferralCode(t *testing.T) {
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
		// clear test data
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

	reqData := struct {
		AppID        string `json:"appID"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        testAppID,
		ReferralCode: testReferralCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/bindReferralCode/v1", &reqData, _testCookie, &respData, nil)
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
	var referralNewUser ReferralNewUser
	if err := _alchemistGormDB.Table(TableNameReferralNewUser).
		Where(`user_id = ? AND app = ? AND referral_code = ? AND deleted_at = 0`,
			_testAccount.ID, testAppID, testReferralCode).
		First(&referralNewUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralNewUser{ID: referralNewUser.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralLog ReferralLog
	if err := _alchemistGormDB.Table(TableNameReferralLog).
		Where(`referral_point_id = ? AND app = ? AND type = ? AND reason = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, consts.ReferralLogTypeGain, consts.ReferralLogReasonNewUserFirstTime, consts.ReferralPointNewUser).
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
	var referralPoint ReferralPoint
	if err := _alchemistGormDB.Table(TableNameReferralPoint).
		Where(`id = ? AND app = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, testPoints+consts.ReferralPointNewUser).
		First(&referralPoint).Error; err != nil {
		t.Error(err)
		return
	}
}

func TestBindReferralCode_NewUserAndNon(t *testing.T) {
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

	// first time +3
	reqData := struct {
		AppID        string `json:"appID"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        testAppID,
		ReferralCode: testReferralCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/bindReferralCode/v1", &reqData, _testCookie, &respData, nil)
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
	var referralNewUser ReferralNewUser
	if err := _alchemistGormDB.Table(TableNameReferralNewUser).
		Where(`user_id = ? AND app = ? AND referral_code = ? AND deleted_at = 0`,
			_testAccount.ID, testAppID, testReferralCode).
		First(&referralNewUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralNewUser{ID: referralNewUser.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralLog ReferralLog
	if err := _alchemistGormDB.Table(TableNameReferralLog).
		Where(`referral_point_id = ? AND app = ? AND type = ? AND reason = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, consts.ReferralLogTypeGain, consts.ReferralLogReasonNewUserFirstTime, consts.ReferralPointNewUser).
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
	var referralPoint ReferralPoint
	if err := _alchemistGormDB.Table(TableNameReferralPoint).
		Where(`id = ? AND app = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, testPoints+consts.ReferralPointNewUser).
		First(&referralPoint).Error; err != nil {
		t.Error(err)
		return
	}

	// second +1, use account2 to bind referral code
	reqData = struct {
		AppID        string `json:"appID"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        testAppID,
		ReferralCode: testReferralCode,
	}
	respData = struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err = postJsonRequest(_kongDNS+"/alchemist/bindReferralCode/v1", &reqData, _testCookie2, &respData, nil)
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
	var referralNewUser2 ReferralNewUser
	if err := _alchemistGormDB.Table(TableNameReferralNewUser).
		Where(`user_id = ? AND app = ? AND referral_code = ? AND deleted_at = 0`,
			_testAccount2.ID, testAppID, testReferralCode).
		First(&referralNewUser2).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralNewUser{ID: referralNewUser2.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralLog2 ReferralLog
	if err := _alchemistGormDB.Table(TableNameReferralLog).
		Where(`referral_point_id = ? AND app = ? AND type = ? AND reason = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, consts.ReferralLogTypeGain, consts.ReferralLogReasonNewUser, consts.ReferralPointNonNewUser).
		First(&referralLog2).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralLog{ID: referralLog2.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	var referralPoint2 ReferralPoint
	if err := _alchemistGormDB.Table(TableNameReferralPoint).
		Where(`id = ? AND app = ? AND points =? AND deleted_at = 0`,
			mockReferralPoint.ID, testAppID, testPoints+consts.ReferralPointNewUser+consts.ReferralPointNonNewUser).
		First(&referralPoint2).Error; err != nil {
		t.Error(err)
		return
	}
}

func TestBindReferralCode_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()

	reqData := struct {
		AppID        string `json:"appID"`
		ReferralCode string `json:"referralCode"`
	}{
		AppID:        testAppID,
		ReferralCode: testReferralCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/bindReferralCode/v1", &reqData, nil, &respData, nil)
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
