package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetLatestSpotPrices(t *testing.T) {
	reqData := struct {
		Symbols  []string `json:"symbols"`
		BaseCoin string   `json:"baseCoin"`
	}{
		Symbols:  []string{"btc", "weth", "eth"},
		BaseCoin: "USDT",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []string `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getLatestSpotPrices/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) <= 0 {
		t.Error("not prospective response data")
		return
	}
}

func TestGetLatestSpotPrices_UnsupportedCryptocurrency(t *testing.T) {
	reqData := struct {
		Symbols  []string `json:"symbols"`
		BaseCoin string   `json:"baseCoin"`
	}{
		Symbols:  []string{"btc12", "weth32", "eth134"},
		BaseCoin: "USDT",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []string `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getLatestSpotPrices/v1", &reqData, nil, &respData, nil)
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
