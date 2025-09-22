package doom_console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-console/api/response"
)

func TestErc20TokensInfo(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: _doomConsoleApiKey,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			HeaderNumber  uint64 `json:"headerNumber,omitempty"`
			ToBlockNumber uint64 `json:"toBlockNumber,omitempty"`
			NumberDiff    uint64 `json:"numberDiff,omitempty"`
			Days          uint32 `json:"days,omitempty"`
			TotalTokens   uint64 `json:"totalTokens,omitempty"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/console/erc20TokensInfo/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.HeaderNumber <= 0 || respData.Data.TotalTokens <= 0 {
		t.Error("not prospective response data data")
		return
	}
}
