package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetAddress(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
	}{
		Address: "bc1qzfnufa2v2rkaaaqapwmsyxs7vq2fu9fq4h9n7h",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Address    string `json:"address,omitempty"`
			ChainStats struct {
				FundedTxoCount int32 `json:"fundedTxoCount,omitempty"`
				FundedTxoSum   int64 `json:"fundedTxoSum,omitempty"`
				SpentTxoCount  int32 `json:"spentTxoCount,omitempty"`
				SpentTxoSum    int64 `json:"spentTxoSum,omitempty"`
				TxCount        int64 `json:"txCount,omitempty"`
			} `json:"chainStats,omitempty"`
			MempoolStats struct {
				FundedTxoCount int32 `json:"fundedTxoCount,omitempty"`
				FundedTxoSum   int64 `json:"fundedTxoSum,omitempty"`
				SpentTxoCount  int32 `json:"spentTxoCount,omitempty"`
				SpentTxoSum    int64 `json:"spentTxoSum,omitempty"`
				TxCount        int32 `json:"txCount,omitempty"`
			} `json:"mempoolStats,omitempty"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/bitcoin/getAddress/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Address == "" {
		t.Error("not prospective response data")
		return
	}
}
