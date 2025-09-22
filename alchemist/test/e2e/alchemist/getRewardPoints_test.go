package alchemist_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/pkg/consts"
	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestGetRewardPoints(t *testing.T) {
	// mock data
	testAppID := "alchemist"

	// insert one referral point
	mockReferralPoint := ReferralPoint{
		UserID: _testAccount.ID,
		App:    testAppID,
		Points: 2,
	}
	if err := _alchemistGormDB.Create(&mockReferralPoint).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _alchemistGormDB.Delete(&ReferralPoint{ID: mockReferralPoint.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	now := time.Now()
	mockReferralLog1 := ReferralLog{
		ReferralPointID: mockReferralPoint.ID,
		App:             testAppID,
		Timestamp:       now.Add(-15 * time.Minute).UnixMilli(),
		Type:            consts.ReferralLogTypeGain,
		Reason:          consts.ReferralLogReasonNewUser,
		Points:          consts.ReferralPointNewUser,
	}
	if err := _alchemistGormDB.Create(&mockReferralLog1).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralLog{ID: mockReferralLog1.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	mockReferralLog2 := ReferralLog{
		ReferralPointID: mockReferralPoint.ID,
		App:             testAppID,
		Timestamp:       now.Add(-10 * time.Minute).UnixMilli(),
		Type:            consts.ReferralLogTypeGain,
		Reason:          consts.ReferralLogReasonNewUser,
		Points:          consts.ReferralPointNonNewUser,
	}
	if err := _alchemistGormDB.Create(&mockReferralLog2).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralLog{ID: mockReferralLog2.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	mockReferralLog3 := ReferralLog{
		ReferralPointID: mockReferralPoint.ID,
		App:             testAppID,
		Timestamp:       now.Add(-5 * time.Minute).UnixMilli(),
		Type:            consts.ReferralLogTypeConsume,
		Reason:          consts.ReferralLogReasonFreeTrial,
		Points:          2,
	}
	if err := _alchemistGormDB.Create(&mockReferralLog3).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&ReferralLog{ID: mockReferralLog3.ID}).Error; err != nil {
			t.Error(err)
			return
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
			Points  int64 `json:"points"`
			Records []struct {
				Timestamp int64  `json:"timestamp"`
				Type      int32  `json:"type"`
				Reason    string `json:"reason"`
				Points    int64  `json:"points"`
			} `json:"records"`
			NumberOfNewUsers    int32 `json:"numberOfNewUsers"`
			NumberOfBilledUsers int32 `json:"numberOfBilledUsers"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getRewardPoints/v1", &reqData, _testCookie, &respData, nil)
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
	if respData.Data.Points != 2 {
		t.Error("not prospective response data")
		return
	}
	if len(respData.Data.Records) != 3 {
		t.Error("not prospective response data")
		return
	}
	if respData.Data.Records[0].Timestamp != now.Add(-5*time.Minute).UnixMilli() &&
		respData.Data.Records[0].Type != consts.ReferralLogTypeConsume {
		t.Error("not prospective response data")
		return
	}
}

func TestGetRewardPoints_EmptySession(t *testing.T) {
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
		Data         struct {
			Points  int64 `json:"points"`
			Records []struct {
				Timestamp int64  `json:"timestamp"`
				Type      int32  `json:"type"`
				Reason    string `json:"reason"`
				Points    int64  `json:"points"`
			} `json:"records"`
			NumberOfNewUsers    int32 `json:"numberOfNewUsers"`
			NumberOfBilledUsers int32 `json:"numberOfBilledUsers"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getRewardPoints/v1", &reqData, nil, &respData, nil)
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
