package doom_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/doom-go/api/response"
)

func TestGetUIKlines_5m(t *testing.T) {
	now := time.Now()
	reqData := struct {
		BeginTime int64  `json:"beginTime"`
		EndTime   int64  `json:"endTime"`
		BaseCoin  string `json:"baseCoin"`
		Symbol    string `json:"symbol"`
		Interval  string `json:"interval"`
	}{
		BeginTime: now.Add(-12 * time.Hour).UnixMilli(),
		EndTime:   now.UnixMilli(),
		BaseCoin:  "USDT",
		Symbol:    "usdc",
		Interval:  "5m",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				OpenTime   int64  `json:"openTime"`
				CloseTime  int64  `json:"closeTime"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getUIKlines/v1", &reqData, nil, &respData, nil)
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

func TestGetUIKlines_1h(t *testing.T) {
	now := time.Now()
	reqData := struct {
		BeginTime int64  `json:"beginTime"`
		EndTime   int64  `json:"endTime"`
		BaseCoin  string `json:"baseCoin"`
		Symbol    string `json:"symbol"`
		Interval  string `json:"interval"`
	}{
		BeginTime: now.Add(7 * -24 * time.Hour).UnixMilli(),
		EndTime:   now.UnixMilli(),
		BaseCoin:  "USDT",
		Symbol:    "usdc",
		Interval:  "1h",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				OpenTime   int64  `json:"openTime"`
				CloseTime  int64  `json:"closeTime"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getUIKlines/v1", &reqData, nil, &respData, nil)
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

func TestGetUIKlines_1d(t *testing.T) {
	now := time.Now()
	reqData := struct {
		BeginTime int64  `json:"beginTime"`
		EndTime   int64  `json:"endTime"`
		BaseCoin  string `json:"baseCoin"`
		Symbol    string `json:"symbol"`
		Interval  string `json:"interval"`
	}{
		BeginTime: now.Add(100 * -24 * time.Hour).UnixMilli(),
		EndTime:   now.UnixMilli(),
		BaseCoin:  "USDT",
		Symbol:    "usdc",
		Interval:  "1d",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				OpenTime   int64  `json:"openTime"`
				CloseTime  int64  `json:"closeTime"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getUIKlines/v1", &reqData, nil, &respData, nil)
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

func TestGetUIKlines_OKX(t *testing.T) {
	now := time.Now()
	reqData := struct {
		BeginTime int64  `json:"beginTime"`
		EndTime   int64  `json:"endTime"`
		BaseCoin  string `json:"baseCoin"`
		Symbol    string `json:"symbol"`
		Interval  string `json:"interval"`
	}{
		BeginTime: now.Add(100 * -24 * time.Hour).UnixMilli(),
		EndTime:   now.UnixMilli(),
		BaseCoin:  "USDT",
		Symbol:    "OKB",
		Interval:  "1d",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				OpenTime   int64  `json:"openTime"`
				CloseTime  int64  `json:"closeTime"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getUIKlines/v1", &reqData, nil, &respData, nil)
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
