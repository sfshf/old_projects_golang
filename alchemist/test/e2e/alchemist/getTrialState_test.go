package alchemist_test

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestGetTrialState(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testTimestamp := time.Now()

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
		// clear test data
		if err := _alchemistGormDB.Delete(&FreeTrialState{ID: mockFreeTrialState.ID}).Error; err != nil {
			log.Println(err)
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
		Data         struct {
			InFreeTrial      bool  `json:"inFreeTrial"`
			ExpirationDate   int64 `json:"expirationDate"`
			StartDate        int64 `json:"startDate"`
			DaysOfTrial      int32 `json:"daysOfTrial"`
			TotalDaysOfTrial int32 `json:"totalDaysOfTrial"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getTrialState/v1", &reqData, _testCookie, &respData, nil)
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
	if !respData.Data.InFreeTrial {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetTrialState_EmptySession(t *testing.T) {
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
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getTrialState/v1", &reqData, nil, &respData, nil)
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
