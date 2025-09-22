package word_test

import (
	"crypto/md5"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/nextsurfer/word/api/response"
)

func TestFetchAudioURL(t *testing.T) {
	// send request
	reqData := struct {
		Text   string `json:"text"`
		Ssml   string `json:"ssml"`
		Accent string `json:"accent"`
		Voice  string `json:"voice"`
		ApiKey string `json:"apiKey"`
	}{
		Text:   "hello aws polly",
		Accent: "us",
		Voice:  "Matthew",
		ApiKey: "75df8c8c2b54d146305b01a9e0d649de",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			AudioURL string `json:"audioURL"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/audio/getAudioURL/v1", &reqData, _testCookie, &respData, nil)
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
	// check whether the url is valid
	resp, err = http.Get(respData.Data.AudioURL)
	if err != nil {
		t.Error(err)
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("Response status from audio url:", resp.StatusCode)
	h := md5.New()
	if _, err := h.Write(data); err != nil {
		t.Error(err)
		return
	}
	log.Printf("Response data from audio url: %x\n", h.Sum(nil))
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response data")
		return
	}
}
