package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetGasFee(t *testing.T) {
	reqData := struct {
		Chain string `json:"chain"`
	}{
		Chain: "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			BaseFee        string `json:"baseFee,omitempty"`
			SlowGasPrice   string `json:"slowGasPrice,omitempty"`
			NormalGasPrice string `json:"normalGasPrice,omitempty"`
			FastGasPrice   string `json:"fastGasPrice,omitempty"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getGasFee/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.BaseFee == "" &&
		respData.Data.SlowGasPrice == "" &&
		respData.Data.NormalGasPrice == "" &&
		respData.Data.FastGasPrice == "" {
		t.Error("not prospective response data")
		return
	}
}
