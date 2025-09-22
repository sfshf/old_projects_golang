package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetTokenApprovals(t *testing.T) {
	testUserAddress := "0x3DdfA8eC3052539b6C9549F12cEA2C295cfF5296" // 36cc7B13029B5DEe4034745FB4F24034f3F2ffc6
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: testUserAddress,
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Approvals []struct {
				Address   string `json:"address,omitempty"`
				Target    string `json:"target,omitempty"`
				Allowance string `json:"allowance,omitempty"`
				Unlimited bool   `json:"unlimited,omitempty"`
				Symbol    string `json:"symbol,omitempty"`
				Name      string `json:"name,omitempty"`
			} `json:"approvals"`
			UnknownApprovals []struct {
				Address   string `json:"address,omitempty"`
				Target    string `json:"target,omitempty"`
				Allowance string `json:"allowance,omitempty"`
				Unlimited bool   `json:"unlimited,omitempty"`
			} `json:"unknownApprovals"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getTokenApprovals/v1", &reqData, nil, &respData, nil)
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
	// check data
	if len(respData.Data.Approvals) <= 0 && len(respData.Data.UnknownApprovals) <= 0 {
		t.Error("not prospective response data")
		return
	}
}
