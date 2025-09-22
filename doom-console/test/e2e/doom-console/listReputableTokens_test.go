package doom_console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-console/api/response"
)

func TestListReputableTokens(t *testing.T) {
	reqData := struct {
		ApiKey     string `json:"apiKey"`
		PageNumber int64  `json:"pageNumber"`
		PageSize   int64  `json:"pageSize"`
	}{
		ApiKey:     _doomConsoleApiKey,
		PageNumber: 0,
		PageSize:   40,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Total int64 `json:"total"`
			List  []struct {
				Address       string   `json:"address,omitempty"`
				Type          string   `json:"type,omitempty"`
				Symbol        string   `json:"symbol,omitempty"`
				Name          string   `json:"name,omitempty"`
				Decimals      uint32   `json:"decimals,omitempty"`
				BinanceSymbol []string `json:"binanceSymbol,omitempty"`
			} `json:"list,omitempty"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/console/listReputableTokens/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) == 0 {
		t.Error("not prospective response data data")
		return
	}
}
