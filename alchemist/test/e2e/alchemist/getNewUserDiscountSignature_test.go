package alchemist_test

import (
	"net/http"
	"testing"

	// . "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

// func TestGetNewUserDiscountSignature(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"
// 	testAppAccountToken := util.NewUUIDString()
// 	testReferralCode := util.GenerateReferralCode()
// 	testTimestamp := time.Now()
// 	testBilledTimes := 3
// 	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

// 	// insert one account token record
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
// 	mockNewUserDiscountState := NewUserDiscountState{
// 		UserID:         _testAccount.ID,
// 		App:            testAppID,
// 		ReferralCode:   testReferralCode,
// 		StartDate:      testTimestamp.UnixMilli(),
// 		BilledTimes:    int32(testBilledTimes),
// 		RemainingTimes: 12 - int32(testBilledTimes),
// 	}
// 	if err := _alchemistGormDB.Create(&mockNewUserDiscountState).Error; err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer func() {
// 		if err := _alchemistGormDB.Delete(&NewUserDiscountState{ID: mockNewUserDiscountState.ID}).Error; err != nil {
// 			t.Error(err)
// 			return
// 		}
// 	}()
// 	reqData := struct {
// 		AppID       string `json:"appID"`
// 		Environment string `json:"environment"`
// 	}{
// 		AppID:       testAppID,
// 		Environment: testEnvironment,
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			OfferID            string `json:"offerID"`
// 			CanConvertToPoints bool   `json:"canConvertToPoints"`
// 			SignatureInfo      struct {
// 				KeyID     string `json:"keyID"`
// 				Nonce     string `json:"nonce"`
// 				Timestamp int64  `json:"timestamp"`
// 				Signature string `json:"signature"`
// 			} `json:"signatureInfo"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/getNewUserDiscountSignature/v1", &reqData, _testCookie, &respData, nil)
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
// 		Where(`offer_id = ? AND user_id = ? AND app = ? AND environment = ? AND deleted_at = 0`,
// 			respData.Data.OfferID, _testAccount.ID, testAppID, consts.EnvironmentNum(testEnvironment)).
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
// 	if !respData.Data.CanConvertToPoints {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestGetNewUserDiscountSignature_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testEnvironment := consts.DESIGNATED_ENVIRONMENT_SANDBOX

	reqData := struct {
		AppID       string `json:"appID"`
		Environment string `json:"environment"`
	}{
		AppID:       testAppID,
		Environment: testEnvironment,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			OfferID            string `json:"offerID"`
			CanConvertToPoints bool   `json:"canConvertToPoints"`
			SignatureInfo      struct {
				KeyID     string `json:"keyID"`
				Nonce     string `json:"nonce"`
				Timestamp int64  `json:"timestamp"`
				Signature string `json:"signature"`
			} `json:"signatureInfo"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getNewUserDiscountSignature/v1", &reqData, nil, &respData, nil)
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
