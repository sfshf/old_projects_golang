package manage_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
)

func TestListConfigs(t *testing.T) {
	// mock data
	testAppID := "alchemist-e2e-TestListConfigs"
	newAppConfig := AppConfig{
		App: testAppID,
		Config: `{
	"appID":"alchemist-e2e-TestListConfigs",
	"productId":"alchemist-e2e",
	"deviceCheck": {
		"keyID": "deviceCheckKeyID",
	"issuerID": "deviceCheckIssuerID",
	"privKeyPem": "deviceCheckPrivKeyPem",
	},
	"discountOffer":{
		"idNewUser": "idNewUser",
	"id10M": "id10M",
	"id8M": "id8M",
	"id6M": "id6M",
	"id4M": "id4M",
	"id2M": "id2M"
	},
	"promoOfferKeyID": "promoOfferKeyID",
	"promoOfferPrivKeyPem": "promoOfferPrivKeyPem",
	"rewardList": [{
		"id": "id",
		"name": "name",
		"description": "description",
		"offerID": "offerID",
		"cost": 9,
		"duration": "duration",
		"durationInDays": 9
	}]
}`,
	}
	if err := _alchemistGormDB.Create(&newAppConfig).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&AppConfig{ID: newAppConfig.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		Password string `json:"password"`
	}{
		Password: _alchemistApiKey,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				ID     int64  `json:"id"`
				Config string `json:"config"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/listConfigs/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) < 1 {
		t.Error("not prospective response data")
		return
	}
}

func TestListConfigs_EmptyParameter(t *testing.T) {
	reqData := struct {
		Password string `json:"password"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/listConfigs/v1", &reqData, nil, &respData, nil)
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
