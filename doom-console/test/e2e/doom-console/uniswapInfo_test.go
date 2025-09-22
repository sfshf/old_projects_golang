package doom_console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-console/api/response"
)

func TestUniswapV2LPInfo(t *testing.T) {
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
			V2Info struct {
				TotalRealTime uint64 `json:"totalRealTime,omitempty"`
				TotalInDB     uint64 `json:"totalInDB,omitempty"`
				DiffValue     uint64 `json:"diffValue,omitempty"`
			} `json:"v2Info,omitempty"`
			V3Info struct {
				Timestamp int64  `json:"timestamp,omitempty"`
				TotalInDB uint64 `json:"totalInDB,omitempty"`
			} `json:"v3Info,omitempty"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/console/uniswapInfo/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.V2Info.TotalRealTime <= 0 || respData.Data.V3Info.Timestamp <= 0 {
		t.Error("not prospective response data data")
		return
	}
}
