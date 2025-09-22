package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetAddressTransactions(t *testing.T) {
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
			List []struct {
				TxID     string `json:"txID,omitempty"`
				Version  int32  `json:"version,omitempty"`
				Locktime int32  `json:"locktime,omitempty"`
				Vin      []struct {
					TxID    string `json:"txID,omitempty"`
					Vout    int32  `json:"vout,omitempty"`
					Prevout struct {
						Scriptpubkey        string `json:"scriptpubkey,omitempty"`
						ScriptpubkeyAsm     string `json:"scriptpubkeyAsm,omitempty"`
						ScriptpubkeyType    string `json:"scriptpubkeyType,omitempty"`
						ScriptpubkeyAddress string `json:"scriptpubkeyAddress,omitempty"`
						Value               int32  `json:"value,omitempty"`
					} `json:"prevout,omitempty"`
					Scriptsig    string   `json:"scriptsig,omitempty"`
					ScriptsigAsm string   `json:"scriptsigAsm,omitempty"`
					Witness      []string `json:"witness,omitempty"`
					IsCoinbase   bool     `json:"isCoinbase,omitempty"`
					Sequence     int64    `json:"sequence,omitempty"`
				} `json:"vin,omitempty"`
				Vout []struct {
					Scriptpubkey        string `json:"scriptpubkey,omitempty"`
					ScriptpubkeyAsm     string `json:"scriptpubkeyAsm,omitempty"`
					ScriptpubkeyType    string `json:"scriptpubkeyType,omitempty"`
					ScriptpubkeyAddress string `json:"scriptpubkeyAddress,omitempty"`
					Value               int32  `json:"value,omitempty"`
				} `json:"vout,omitempty"`
				Size   int32 `json:"size,omitempty"`
				Weight int32 `json:"weight,omitempty"`
				Fee    int32 `json:"fee,omitempty"`
				Status struct {
					Confirmed   bool   `json:"confirmed,omitempty"`
					BlockHeight int32  `json:"blockHeight,omitempty"`
					BlockHash   string `json:"blockHash,omitempty"`
					BlockTime   int64  `json:"blockTime,omitempty"`
				} `json:"status,omitempty"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/bitcoin/getAddressTransactions/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) <= 0 {
		t.Error("not prospective response data")
		return
	}
}
