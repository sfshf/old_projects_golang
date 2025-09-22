package alchemist_test

import (
	"net/http"
	"testing"

	// . "github.com/nextsurfer/alchemist/internal/pkg/model"
	slark_response "github.com/nextsurfer/slark/api/response"
)

// func TestRedeemReward(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"
// 	testReferralPoints := 100
// 	testRewardID := "promo.trial.6m"
// 	testRewardCost := 56
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
// 		if err := _alchemistGormDB.Delete(&ReferralPoint{ID: mockReferralPoint.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	mockFreeTrialState := FreeTrialState{
// 		App:            testAppID,
// 		UserID:         _testAccount.ID,
// 		ExpirationDate: testTimestamp.AddDate(0, 0, -1).UnixMilli(),
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
// 	reqData := struct {
// 		AppID    string `json:"appID"`
// 		RewardID string `json:"rewardID"`
// 	}{
// 		AppID:    testAppID,
// 		RewardID: testRewardID,
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			RemainingPoints int32 `json:"remainingPoints"`
// 			ExpirationDate  int64 `json:"expirationDate"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/redeemReward/v1", &reqData, _testCookie, &respData, nil)
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
// 		Where(`referral_point_id = ? AND app = ? AND type = ? AND reason = ? AND points = ? AND deleted_at = 0`,
// 			mockReferralPoint.ID, testAppID, consts.ReferralLogTypeConsume, consts.ReferralLogReasonRedeemReward, testRewardCost).
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
// 	var freeTrialState FreeTrialState
// 	if err := _alchemistGormDB.Table(TableNameFreeTrialState).
// 		Where(`id = ? AND user_id = ? AND app = ? AND deleted_at = 0`,
// 			mockFreeTrialState.ID, _testAccount.ID, testAppID).
// 		First(&freeTrialState).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&FreeTrialState{ID: freeTrialState.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	if !time.UnixMilli(freeTrialState.ExpirationDate).After(testTimestamp) ||
// 		freeTrialState.DaysOfTrial <= 0 ||
// 		freeTrialState.TotalDaysOfTrial <= 0 {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestRedeemReward_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testRewardID := "promo.trial.6m"

	reqData := struct {
		AppID    string `json:"appID"`
		RewardID string `json:"rewardID"`
	}{
		AppID:    testAppID,
		RewardID: testRewardID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			RemainingPoints int32 `json:"remainingPoints"`
			ExpirationDate  int64 `json:"expirationDate"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/redeemReward/v1", &reqData, nil, &respData, nil)
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
