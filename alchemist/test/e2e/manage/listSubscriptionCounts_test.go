package manage_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
)

func TestListSubscriptionCounts(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	loc := time.FixedZone("UTC-5", -5*60*60)
	utc5 := time.Now().In(loc)
	testEndTS, _ := time.ParseInLocation("2006-01-02", utc5.Format("2006-01-02"), loc)
	testStartTS := testEndTS.AddDate(0, 0, -1)

	newSubscriptionCount1 := SubscriptionCount{
		App:   testAppID,
		Time:  testStartTS.UnixMilli(),
		Count: 1,
	}
	if err := _alchemistGormDB.Create(&newSubscriptionCount1).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&SubscriptionCount{ID: newSubscriptionCount1.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	newSubscriptionCount2 := SubscriptionCount{
		App:   testAppID,
		Time:  testEndTS.UnixMilli(),
		Count: 2,
	}
	if err := _alchemistGormDB.Create(&newSubscriptionCount2).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&SubscriptionCount{ID: newSubscriptionCount2.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		Password  string `json:"password"`
		AppID     string `json:"appID"`
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}{
		Password:  _alchemistApiKey,
		AppID:     testAppID,
		StartDate: testStartTS.Format("2006-01-02"),
		EndDate:   testEndTS.Format("2006-01-02"),
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []int64 `json:"list"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/listSubscriptionCounts/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) != 2 {
		t.Error("not prospective response data")
		return
	}
}

func TestListSubscriptionCounts_EmptyParameter(t *testing.T) {
	reqData := struct {
		Password  string `json:"password"`
		App       string `json:"app"`
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/listSubscriptionCounts/v1", &reqData, nil, &respData, nil)
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
