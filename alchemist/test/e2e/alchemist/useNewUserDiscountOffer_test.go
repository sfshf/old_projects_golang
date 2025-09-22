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

func TestUseNewUserDiscountOffer(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()
	testTimestamp := time.Now()
	testOfferID := _DiscountOfferID8M
	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

	mockReferralCode := ReferralCode{
		App:          testAppID,
		UserID:       _testAccount.ID,
		JoinDate:     testTimestamp,
		ReferralCode: testReferralCode,
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
	mockPromoOfferRecord := PromoOfferRecord{
		UserID:      _testAccount.ID,
		App:         testAppID,
		OfferID:     testOfferID,
		Environment: consts.EnvironmentNum(testEnvironment),
	}
	if err := _alchemistGormDB.Create(&mockPromoOfferRecord).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&PromoOfferRecord{ID: mockPromoOfferRecord.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		AppID       string `json:"appID"`
		OfferID     string `json:"offerID"`
		Environment string `json:"environment"`
	}{
		AppID:       testAppID,
		OfferID:     testOfferID,
		Environment: testEnvironment,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/useNewUserDiscountOffer/v1", &reqData, _testCookie, &respData, nil)
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
	var newUserDiscountState NewUserDiscountState
	if err := _alchemistGormDB.Table(TableNameNewUserDiscountState).
		Where(`user_id = ? AND app = ? AND referral_code = ? AND deleted_at = 0`,
			_testAccount.ID, testAppID, testReferralCode).
		First(&newUserDiscountState).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&NewUserDiscountState{ID: newUserDiscountState.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
}

func TestUseNewUserDiscountOffer_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testOfferID := _DiscountOfferID8M
	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

	reqData := struct {
		AppID       string `json:"appID"`
		OfferID     string `json:"offerID"`
		Environment string `json:"environment"`
	}{
		AppID:       testAppID,
		OfferID:     testOfferID,
		Environment: testEnvironment,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/useNewUserDiscountOffer/v1", &reqData, nil, &respData, nil)
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
