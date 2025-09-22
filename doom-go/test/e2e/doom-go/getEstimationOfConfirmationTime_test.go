package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetEstimationOfConfirmationTime(t *testing.T) {
	reqData := struct {
		Chain    string `json:"chain"`
		GasPrice string `json:"gasPrice"`
	}{
		Chain:    "eth",
		GasPrice: "200000000",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			EstimatedSeconds string `json:"estimatedSeconds,omitempty"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getEstimationOfConfirmationTime/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.EstimatedSeconds == "" {
		t.Error("not prospective response data")
		return
	}
}
