package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetABI_Normal_NotProxy(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ABI           string `json:"ABI"`
			IsProxy       bool   `json:"isProxy"`
			TargetAddress string `json:"targetAddress"`
			ProxyType     string `json:"proxyType"`
			ProxyABI      string `json:"proxyABI"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getABI/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.ABI == "" || respData.Data.IsProxy {
		t.Error("not prospective response data")
		return
	}
}

func TestGetABI_Unnormal_NotProxy(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "0x518Ae805Bd145c8Ed1e22EFD0b21bad253Cf1BED",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ABI           string `json:"ABI"`
			IsProxy       bool   `json:"isProxy"`
			TargetAddress string `json:"targetAddress"`
			ProxyType     string `json:"proxyType"`
			ProxyABI      string `json:"proxyABI"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getABI/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeBadRequest {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetABI_Normal_Proxy1(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "A0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ABI           string `json:"ABI"`
			IsProxy       bool   `json:"isProxy"`
			TargetAddress string `json:"targetAddress"`
			ProxyType     string `json:"proxyType"`
			ProxyABI      string `json:"proxyABI"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getABI/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.ABI == "" || !respData.Data.IsProxy {
		t.Error("not prospective response data")
		return
	}
}

func TestGetABI_Normal_Proxy2(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "114f1388fAB456c4bA31B1850b244Eedcd024136",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ABI           string `json:"ABI"`
			IsProxy       bool   `json:"isProxy"`
			TargetAddress string `json:"targetAddress"`
			ProxyType     string `json:"proxyType"`
			ProxyABI      string `json:"proxyABI"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getABI/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.ABI == "" || !respData.Data.IsProxy {
		t.Error("not prospective response data")
		return
	}
}
