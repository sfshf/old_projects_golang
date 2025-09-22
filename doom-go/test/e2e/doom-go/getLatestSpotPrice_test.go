package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetLatestSpotPrice(t *testing.T) {
	reqData := struct {
		Symbol   string `json:"symbol"`
		BaseCoin string `json:"baseCoin"`
	}{
		Symbol:   "BTC",
		BaseCoin: "USDT",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Price string `json:"price"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getLatestSpotPrice/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Price == "" {
		t.Error("not prospective response data")
		return
	}
}

func TestGetLatestSpotPrice_UnsupportedCryptocurrency(t *testing.T) {
	reqData := struct {
		Symbol   string `json:"symbol"`
		BaseCoin string `json:"baseCoin"`
	}{
		Symbol:   "FAKEBTC",
		BaseCoin: "USDT",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Price string `json:"price"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getLatestSpotPrice/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeUnsupportedCryptocurrency {
		t.Error("not prospective response data code")
		return
	}
}
