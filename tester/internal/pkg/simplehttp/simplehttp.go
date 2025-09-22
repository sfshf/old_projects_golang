package simplehttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	ErrResponseStatusCodeNotEqualTo200 = errors.New("http response status code not equal to 200")
	ErrResponseDataCodeNotEqualToZero  = errors.New("http response data code not equal to 0")
)

func Get(url string, headers map[string]string, respData interface{}) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if respData != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, respData); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func PostJsonRequest(location string, headers map[string]string, reqData interface{}, cookie *http.Cookie, respData interface{}) (*http.Response, error) {
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(http.MethodPost, location, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if respData != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, respData); err != nil {
			return resp, err
		}
	}
	return resp, nil
}
