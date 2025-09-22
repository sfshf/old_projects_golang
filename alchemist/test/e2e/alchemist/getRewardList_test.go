package alchemist_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
)

// func TestGetRewardList(t *testing.T) {
// 	// mock data
// 	testAppID := "alchemist"

// 	reqData := struct {
// 		AppID string `json:"appID"`
// 	}{
// 		AppID: testAppID,
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			RewardList []struct {
// 				Id             string `json:"id"`
// 				Name           string `json:"name"`
// 				Description    string `json:"description"`
// 				OfferID        string `json:"offerID"`
// 				Cost           int32  `json:"cost"`
// 				Duration       string `json:"duration"`
// 				DurationInDays int32  `json:"durationInDays"`
// 			} `json:"rewardList"`
// 		} `json:"data"`
// 	}{}
// 	// send request
// 	resp, err := postJsonRequest(_kongDNS+"/alchemist/getRewardList/v1", &reqData, _testCookie, &respData, nil)
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
// 	if len(respData.Data.RewardList) <= 0 {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestGetRewardList_EmptyParameter(t *testing.T) {
	reqData := struct {
		AppID string `json:"appID"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/getRewardList/v1", &reqData, nil, &respData, nil)
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
