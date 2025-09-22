package alchemist_test

import (
	"net/http"
	"testing"

	// . "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

// func TestFinishRewardOffer(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"
// 	testReferralPoints := 100
// 	testRewardID := "promo.trial.6m"
// 	testCost := 56
// 	testOfferID := "promo.trial.6m"
// 	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX
// 	testTimestamp := time.Now()

// 	mockReferralPoint := ReferralPoint{
// 		App:    testAppID,
// 		UserID: _testAccount.ID,
// 		Points: int32(testReferralPoints),
// 	}
// 	if err := _alchemistGormDB.Create(&mockReferralPoint).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		// clear test data
// 		if err := _alchemistGormDB.Delete(&ReferralPoint{ID: mockReferralPoint.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	mockFreeTrialState := FreeTrialState{
// 		App:              testAppID,
// 		UserID:           _testAccount.ID,
// 		StartDate:        testTimestamp.UnixMilli(),
// 		ExpirationDate:   testTimestamp.AddDate(0, 0, 1).UnixMilli(),
// 		DaysOfTrial:      1,
// 		TotalDaysOfTrial: 1,
// 	}
// 	if err := _alchemistGormDB.Create(&mockFreeTrialState).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&FreeTrialState{ID: mockFreeTrialState.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	mockPromoOfferRecord := PromoOfferRecord{
// 		App:         testAppID,
// 		UserID:      _testAccount.ID,
// 		OfferID:     testOfferID,
// 		Environment: consts.EnvironmentNum(testEnvironment),
// 	}
// 	if err := _alchemistGormDB.Create(&mockPromoOfferRecord).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&PromoOfferRecord{ID: mockPromoOfferRecord.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()

// 	reqData := struct {
// 		AppID       string `json:"appID"`
// 		RewardID    string `json:"rewardID"`
// 		Environment string `json:"environment"`
// 	}{
// 		AppID:       testAppID,
// 		RewardID:    testRewardID,
// 		Environment: testEnvironment,
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			RemainingPoints int32 `json:"remainingPoints"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/finishRewardOffer/v1", &reqData, _testCookie, &respData, nil)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		t.Error("not prospective response code")
// 		return
// 	}
// 	if respData.Code != response.StatusCodeOK {
// 		t.Error("not prospective response data code")
// 		return
// 	}
// 	// check data
// 	var referralLog ReferralLog
// 	if err := _alchemistGormDB.Table(TableNameReferralLog).
// 		Where(`app = ? AND referral_point_id = ? AND type = ? AND reason = ? AND points = ? AND deleted_at = 0`,
// 			testAppID, mockReferralPoint.ID, consts.ReferralLogTypeConsume, consts.ReferralLogReasonRedeemReward, testCost).
// 		First(&referralLog).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&ReferralLog{ID: referralLog.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	var referralPoint ReferralPoint
// 	if err := _alchemistGormDB.Table(TableNameReferralPoint).
// 		Where(`id = ? AND deleted_at = 0`, mockReferralPoint.ID).
// 		First(&referralPoint).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if referralPoint.Points != mockReferralPoint.Points-int32(testCost) {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestFinishRewardOffer_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testRewardID := "promo.trial.6m"
	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

	reqData := struct {
		AppID       string `json:"appID"`
		RewardID    string `json:"rewardID"`
		Environment string `json:"environment"`
	}{
		AppID:       testAppID,
		RewardID:    testRewardID,
		Environment: testEnvironment,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			RemainingPoints int32 `json:"remainingPoints"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/finishRewardOffer/v1", &reqData, nil, &respData, nil)
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
