package alchemist_test

import (
	"net/http"
	"testing"

	// . "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

// func TestRedeemRewardOffer(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"
// 	testAppAccountToken := util.NewUUIDString()
// 	testReferralPoints := 100
// 	testRewardID := "promo.trial.6m"
// 	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

// 	mockAccountToken := SlarkUser{
// 		AppAccountToken: testAppAccountToken,
// 		UserID:          _testAccount.ID,
// 	}
// 	if err := _alchemistGormDB.Create(&mockAccountToken).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&SlarkUser{ID: mockAccountToken.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()

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
// 			OfferID       string `json:"offerID"`
// 			SignatureInfo struct {
// 				KeyID     string `json:"keyID"`
// 				Nonce     string `json:"nonce"`
// 				Timestamp int64  `json:"timestamp"`
// 				Signature string `json:"signature"`
// 			} `json:"signatureInfo"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/redeemRewardOffer/v1", &reqData, _testCookie, &respData, nil)
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
// 	var promoOfferRecord PromoOfferRecord
// 	if err := _alchemistGormDB.Table(TableNamePromoOfferRecord).
// 		Where(`app = ? AND user_id = ? AND offer_id = ? AND environment = ? AND deleted_at = 0`,
// 			testAppID, _testAccount.ID, respData.Data.OfferID, consts.EnvironmentNum(testEnvironment)).
// 		First(&promoOfferRecord).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&PromoOfferRecord{ID: promoOfferRecord.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// }

func TestRedeemRewardOffer_EmptySession(t *testing.T) {
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
			OfferID       string `json:"offerID"`
			SignatureInfo struct {
				KeyID     string `json:"keyID"`
				Nonce     string `json:"nonce"`
				Timestamp int64  `json:"timestamp"`
				Signature string `json:"signature"`
			} `json:"signatureInfo"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/redeemRewardOffer/v1", &reqData, nil, &respData, nil)
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
