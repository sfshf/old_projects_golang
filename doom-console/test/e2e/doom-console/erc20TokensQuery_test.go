package doom_console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-console/api/response"
)

func TestERC20TokensQuery(t *testing.T) {
	testAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	reqData := struct {
		ApiKey          string `json:"apiKey"`
		ContractAddress string `json:"contractAddress"`
	}{
		ApiKey:          _doomConsoleApiKey,
		ContractAddress: testAddress,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Address  string `json:"address,omitempty"`
			Type     string `json:"type,omitempty"`
			Name     string `json:"name,omitempty"`
			Symbol   string `json:"symbol,omitempty"`
			Decimals uint32 `json:"decimals,omitempty"`
			Priced   bool   `json:"priced,omitempty"`
			Checked  bool   `json:"checked,omitempty"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/console/erc20TokensQuery/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Type == "" {
		t.Error("not prospective response data data")
		return
	}
}

func TestERC20TokensQuery_EmptyAddress(t *testing.T) {
	reqData := struct {
		ApiKey          string `json:"apiKey"`
		ContractAddress string `json:"contractAddress"`
	}{
		ApiKey:          _doomConsoleApiKey,
		ContractAddress: "",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Address  string `json:"address,omitempty"`
			Type     string `json:"type,omitempty"`
			Name     string `json:"name,omitempty"`
			Symbol   string `json:"symbol,omitempty"`
			Decimals uint32 `json:"decimals,omitempty"`
			Priced   bool   `json:"priced,omitempty"`
			Checked  bool   `json:"checked,omitempty"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/console/erc20TokensQuery/v1", &reqData, nil, &respData, nil)
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
