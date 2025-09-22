package alchemist_test

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestGetNewUserDiscountState(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()
	testTimestamp := time.Now()
	testBilledTimes := 3

	// insert new-user discount state record
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
		// clear test data
		if err := _alchemistGormDB.Delete(&NewUserDiscountState{ID: mockNewUserDiscountState.ID}).Error; err != nil {
			log.Println(err)
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
			HasNewUserDiscount bool  `json:"hasNewUserDiscount"`
			Redeemed           bool  `json:"redeemed"`
			RemainingTimes     int32 `json:"remainingTimes"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getNewUserDiscountState/v1", &reqData, _testCookie, &respData, nil)
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
	if !respData.Data.HasNewUserDiscount {
		t.Error("not prospective response data")
		return
	}
}

func TestGetNewUserDiscountState_EmptySession(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testBilledTimes := 3

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
			HasNewUserDiscount bool  `json:"hasNewUserDiscount"`
			Redeemed           bool  `json:"redeemed"`
			RemainingTimes     int32 `json:"remainingTimes"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getNewUserDiscountState/v1", &reqData, nil, &respData, nil)
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
