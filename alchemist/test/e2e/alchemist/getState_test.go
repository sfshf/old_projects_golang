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

func TestGetState(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()
	testTimestamp := time.Now()
	testBilledTimes := 3

	// referral code record
	mockReferralCode := ReferralCode{
		UserID:       _testAccount.ID,
		App:          testAppID,
		JoinDate:     testTimestamp,
		ReferralCode: util.GenerateReferralCode(),
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
	// user_registered_on_old_device record
	mockUserRegisteredOnOldDevice := UserRegisteredOnOldDevice{
		App:          testAppID,
		UserID:       _testAccount.ID,
		IP:           getLocalIPv4(),
		ReferralCode: mockReferralCode.ReferralCode,
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
	// slark_user record
	mockSlarkUser := SlarkUser{
		AppAccountToken: util.NewUUIDString(),
		RegisteredAt:    testTimestamp.UnixMilli(),
		UserID:          _testAccount.ID,
	}
	if err := _alchemistGormDB.Create(&mockSlarkUser).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&SlarkUser{ID: mockSlarkUser.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	// referral points
	mockReferralPoint := &ReferralPoint{
		App:    testAppID,
		UserID: _testAccount.ID,
		Points: 10,
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
	// free trial state
	mockFreeTrialState := FreeTrialState{
		App:              testAppID,
		UserID:           _testAccount.ID,
		ExpirationDate:   testTimestamp.AddDate(0, 0, 1).UnixMilli(),
		StartDate:        testTimestamp.UnixMilli(),
		DaysOfTrial:      1,
		TotalDaysOfTrial: 1,
	}
	if err := _alchemistGormDB.Create(&mockFreeTrialState).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&FreeTrialState{ID: mockFreeTrialState.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	// new user discount state
	mockNewUserDiscountState := NewUserDiscountState{
		UserID:         _testAccount.ID,
		App:            testAppID,
		ReferralCode:   testReferralCode,
		StartDate:      testTimestamp.UnixMilli(),
		BilledTimes:    int32(testBilledTimes),
		RemainingTimes: 12 - int32(testBilledTimes),
	}
	if err := _alchemistGormDB.Create(&mockNewUserDiscountState).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&NewUserDiscountState{ID: mockNewUserDiscountState.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		AppID       string `json:"appID"`
		BilledTimes int    `json:"billedTimes"`
	}{
		AppID:       testAppID,
		BilledTimes: testBilledTimes,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ReferralCodeState struct {
				MyReferralCode              *string `json:"myReferralCode"`
				NewUserStatusExpirationDate *int64  `json:"newUserStatusExpirationDate"`
				ShareURL                    string  `json:"shareURL"`
				UsedReferralCode            *string `json:"usedReferralCode"`
			} `json:"referralCodeState"`
			RewardPoints int32 `json:"rewardPoints"`
			TrialState   struct {
				InFreeTrial      bool  `json:"inFreeTrial"`
				ExpirationDate   int64 `json:"expirationDate"`
				StartDate        int64 `json:"startDate"`
				DaysOfTrial      int32 `json:"daysOfTrial"`
				TotalDaysOfTrial int32 `json:"totalDaysOfTrial"`
			} `json:"trialState"`
			NewUserDiscountState struct {
				HasNewUserDiscount bool  `json:"hasNewUserDiscount"`
				Redeemed           bool  `json:"redeemed"`
				RemainingTimes     int32 `json:"remainingTimes"`
			} `json:"newUserDiscountState"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getState/v1", &reqData, _testCookie, &respData, nil)
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
	if !respData.Data.TrialState.InFreeTrial {
		t.Error("not prospective response data")
		return
	}
}

func TestGetState_EmptySession(t *testing.T) {
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
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getState/v1", &reqData, nil, &respData, nil)
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
