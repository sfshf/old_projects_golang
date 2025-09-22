package doom_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestJsEncrypt(t *testing.T) {
	testPlainText := `{"question1":"asdf","question2":"asdf","question3":"asdf","encryptedPassword":"X2A1vQGH8IH0r9G/rUNdmw+6dms=","hashOfHash":"37ffdc7a8c2bc765a85b1c41965f81f859f352d29d3a6d9f6ea5353e25fcc336","nonce":"+3UVI3qqaZet3AS3ZJcrlAFC+atWimIS"}`
	base64PlainText := base64.StdEncoding.EncodeToString([]byte(testPlainText))
	log.Println("base64PlainText:", base64PlainText)
	testCiphertext := `BNvA7AJFaBIiBc0gGgAHUUb2LL82wynqZgm+pDELjvY+G9xDPcKxu0vBZlsFaKLoKJVcg2rEw3eNVfL9Gllxyti0ZbznmLB6d/bB8WNkgIBWpvVFYE5HAMkAKPfzIAVS7HphR7g3SE8p2vW1M8PF0FFX5OjDPraEogYUk2wnzSVAXpPf3r5g9nSgC3eYrbI71RRUXOkd1WwuAFF+ReCUgONwTyvt1PvqaXrG2Zbv4dRNHR9pxg5HfSK+Js8s2ZPbo2alVPXamcUkeha5oXY4dojwU5GrvBMRGELVyFYd/dVTJ82N7LjfWYb+rUP86SVvHYrz+BLh18n6yKMPfHVZEefA+h3+gfdnwWPTgxqx5U72/mfES2oUEzkD7WGyMiRjXn6Ys17iZBHrxadq6xAEpuRje+OBCHaxmITVFhsT34QwOkOnHuJjfvbwUhJCMzCWNCHuknIjhTUI4rHEHltj/MNZxZIoi93dWVHeqgkJLyvUQe8DKz2pcH7tBC5HdPLarVC4sI62hSeCE1PR8PxzGq0NGQ3QjsC5yw==`
	testTitle := "TestJsEncrypt"
	// TestCreateData ------------------------------------
	reqDataCreateData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{
		PlainText:  base64PlainText,
		CipherText: testCiphertext,
		Title:      testTitle,
	}
	respDataCreateData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respCreateData, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqDataCreateData, _testCookie, &respDataCreateData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}

	defer func() {
		// TestDeleteData ------------------------------------
		reqDataDeleteData := struct {
			Title string `json:"title"`
		}{
			Title: testTitle,
		}
		respDataDeleteData := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeleteData, err := postJsonRequest(_kongDNS+"/doom/deleteData/v1", &reqDataDeleteData, _testCookie, &respDataDeleteData, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respDeleteData.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeleteData.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		// delete mysql mock data
		if _, err := _mongoDB.Collection(CollectionName_Datum).
			DeleteMany(context.Background(), bson.D{{Key: "userID", Value: _testAccount.ID},
				{Key: "dataID", Value: fmt.Sprintf("%d-data-%s", _testAccount.ID, testTitle)}}); err != nil {
			t.Error(err)
		}
	}()

	// TestGetData ------------------------------------
	reqDataGetData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetData, err := postJsonRequest(_kongDNS+"/doom/getData/v1", &reqDataGetData, _testCookie, &respDataGetData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataGetData.Data.PlainText != base64PlainText {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqData, _testCookie, &respData, nil)
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
	if respData.Message != "Request has wrong parameters, please inspect parameters" {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqData, _testCookie, &respData, func(req *http.Request) {
		req.Header.Set("Accept-Language", "zh")
	})
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
	if respData.Message != "请求参数有误，请您排错后重试" {
		t.Error("not prospective response data")
		return
	}
}
