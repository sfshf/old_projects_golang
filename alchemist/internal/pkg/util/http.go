package util

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func PostJsonRequest(location string, reqData interface{}, cookie *http.Cookie, respData interface{}) (*http.Response, error) {
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(http.MethodPost, location, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if respData == nil {
		return resp, nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("resp: %#v\n", resp)
		log.Println(err)
		return nil, err
	}
	if err := json.Unmarshal(data, respData); err != nil {
		log.Printf("respData: %s\n", data)
		log.Println(err)
		return resp, err
	}
	return resp, nil
}
