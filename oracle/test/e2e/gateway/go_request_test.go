package gateway_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/simplecrypto"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestGetLatestSpotPrice(t *testing.T) {
	data := struct {
		Chain    string `json:"chain"`
		GasPrice string `json:"gasPrice"`
	}{
		Chain:    "eth",
		GasPrice: "200000000",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
		return
	}
	ts := time.Now().UnixMilli()
	goKey, err := simplecrypto.Keccak256([]byte(fmt.Sprintf("%d9C9B913EB1B6254F4737CE947", ts)))
	if err != nil {
		t.Error(err)
		return
	}
	aead, err := chacha20poly1305.NewX(goKey)
	if err != nil {
		t.Error(err)
		return
	}
	reqData := struct {
		Path      string `json:"path"`
		Data      string `json:"data"`
		Timestamp int64  `json:"timestamp"`
	}{
		Path:      "/doom/getEstimationOfConfirmationTime/v1",
		Data:      string(simplecrypto.EncryptByX(dataStr, aead, nil, simplecrypto.NonceZeroX())),
		Timestamp: ts,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			EncryptedData interface{} `json:"encryptedData"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_oracleGwDNS+"/go", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != 0 {
		t.Error("not prospective response data code")
		return
	}
	if respData.Data.EncryptedData == nil {
		t.Error("not prospective response data")
		return
	}
	decryptedData, err := simplecrypto.DecryptByX([]byte(respData.Data.EncryptedData.(string)), aead, simplecrypto.NonceZeroX())
	if err != nil {
		t.Error(err)
		return
	}
	var result struct {
		EstimatedSeconds string `json:"estimatedSeconds,omitempty"`
	}
	if err := json.Unmarshal(decryptedData, &result); err != nil {
		t.Error(err)
		return
	}
	if result.EstimatedSeconds == "" {
		t.Error("not prospective response data")
		return
	}
}

func TestGetLatestSpotPrice_TimestampDelay(t *testing.T) {
	data := struct {
		Symbol   string `json:"symbol"`
		BaseCoin string `json:"baseCoin"`
	}{
		Symbol:   "BTC",
		BaseCoin: "USDT",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
		return
	}
	ts := time.Now().Add(-61 * time.Second).UnixMilli()
	goKey, err := simplecrypto.Keccak256([]byte(fmt.Sprintf("%d9C9B913EB1B6254F4737CE947", ts)))
	if err != nil {
		t.Error(err)
		return
	}
	aead, err := chacha20poly1305.NewX(goKey)
	if err != nil {
		t.Error(err)
		return
	}
	reqData := struct {
		Path      string `json:"path"`
		Data      string `json:"data"`
		Timestamp int64  `json:"timestamp"`
	}{
		Path:      "/doom/getLatestSpotPrice/v1",
		Data:      string(simplecrypto.EncryptByX(dataStr, aead, nil, simplecrypto.NonceZeroX())),
		Timestamp: ts,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			EncryptedData interface{} `json:"encryptedData"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_oracleGwDNS+"/go", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeGoRequestDelay {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetLatestSpotPrice_InvalidEncryptKey(t *testing.T) {
	data := struct {
		Symbol   string `json:"symbol"`
		BaseCoin string `json:"baseCoin"`
	}{
		Symbol:   "BTC",
		BaseCoin: "USDT",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
		return
	}
	ts := time.Now()
	goKey, err := simplecrypto.Keccak256([]byte(fmt.Sprintf("%d9C9B913EB1B6254F4737CE947", ts.UnixMilli())))
	if err != nil {
		t.Error(err)
		return
	}
	aead, err := chacha20poly1305.NewX(goKey)
	if err != nil {
		t.Error(err)
		return
	}
	reqData := struct {
		Path      string `json:"path"`
		Data      string `json:"data"`
		Timestamp int64  `json:"timestamp"`
	}{
		Path:      "/doom/getLatestSpotPrice/v1",
		Data:      string(simplecrypto.EncryptByX(dataStr, aead, nil, simplecrypto.NonceZeroX())),
		Timestamp: ts.Add(-30 * time.Second).UnixMilli(),
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			EncryptedData interface{} `json:"encryptedData"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_oracleGwDNS+"/go", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeGoRequestDeformedSecretKey {
		t.Error("not prospective response data code")
		return
	}
}
