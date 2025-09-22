package admin_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_oracleMysqlDsn = os.Getenv("CONSOLE_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_oracleGormDB   *gorm.DB

	_oracleConsoleDNS = "http://172.31.29.192:8866"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("CONSUL_HTTP_ADDR", "172.31.29.192:8500")
	var err error
	// mysql
	_oracleGormDB, err = gorm.Open(mysql.Open(_oracleMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func postJsonRequest(location string, reqData interface{}, cookie *http.Cookie, respData interface{}, reqHeaderFunc func(req *http.Request)) (*http.Response, error) {
	log.Printf("location: %s\n", location)
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return nil, err
		}
		log.Printf("request data: %s\n", jsonData)
		body = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(http.MethodPost, location, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if reqHeaderFunc != nil {
		reqHeaderFunc(req)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("response data: %s\n", data)
	if respData != nil {
		if err := json.Unmarshal(data, respData); err != nil {
			return resp, err
		}
	}
	return resp, nil
}
