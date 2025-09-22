package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetAssets(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "0x3DdfA8eC3052539b6C9549F12cEA2C295cfF5296",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ToBlock string `json:"toBlock"`
			Tokens  []struct {
				Balance string `json:"balance"`
				Address string `json:"address"`
				Name    string `json:"name"`
				Symbol  string `json:"symbol"`
				Price   string `json:"price"`
				Value   string `json:"value"`
			} `json:"tokens"`
			UnknownTokens []struct {
				Balance string `json:"balance"`
				Address string `json:"address"`
			} `json:"unknownTokens"`
			Balance struct {
				Balance string `json:"balance"`
				Price   string `json:"price"`
				Value   string `json:"value"`
			} `json:"balance"`
			DappAssets []struct {
				App   string `json:"app"`
				Value string `json:"value"`
			} `json:"dappAssets"`
			TotalValue string `json:"totalValue"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getAssets/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.Tokens) <= 0 &&
		len(respData.Data.UnknownTokens) <= 0 &&
		len(respData.Data.DappAssets) <= 0 {
		t.Error("not prospective response data")
		return
	}
}

func TestGetAssets_EmptyParameter(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: "",
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Tokens []struct {
				Balance string `json:"balance"`
				Address string `json:"address"`
				Name    string `json:"name"`
				Symbol  string `json:"symbol"`
				Price   string `json:"price"`
				Value   string `json:"value"`
			} `json:"tokens"`
			DappAssets []struct {
				App   string `json:"app"`
				Value string `json:"value"`
			} `json:"dappAssets"`
			TotalValue string `json:"totalValue"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getAssets/v1", &reqData, nil, &respData, nil)
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
