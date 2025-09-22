package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetBalance(t *testing.T) {
	testUserAddress := "3DdfA8eC3052539b6C9549F12cEA2C295cfF5296" // 36cc7B13029B5DEe4034745FB4F24034f3F2ffc6
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: testUserAddress,
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Balance string `json:"balance"`
			Price   string `json:"price"`
			Value   string `json:"value"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getBalance/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Price == "" && respData.Data.Balance == "" && respData.Data.Value == "" {
		t.Error("not prospective response data")
		return
	}
}
