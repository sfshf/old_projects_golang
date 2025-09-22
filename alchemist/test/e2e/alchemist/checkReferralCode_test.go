package alchemist_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
)

func TestCheckReferralCode(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()

	// insert one referral code record
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

	reqData := struct {
		ReferralCode string `json:"referralCode"`
	}{
		ReferralCode: testReferralCode,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/checkReferralCode/v1", &reqData, nil, &respData, nil)
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
	if !respData.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}

func TestCheckReferralCode_EmptyParameter(t *testing.T) {
	// mock data
	reqData := struct {
		ReferralCode string `json:"referralCode"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/checkReferralCode/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
