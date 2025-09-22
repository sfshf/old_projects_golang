package connector_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_kongDNS = "https://api.test.n1xt.net"

	_connectorMysqlDsn = os.Getenv("CONNECTOR_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_connectorGormDB   *gorm.DB
	_connectorRedisDsn = os.Getenv("CONNECTOR_REDIS_DNS")
	_redisCli          *redis.Client

	_adminApiKey = "/4pKidQqVz+SQ9G8G1pPJVb2" // admin
)

func TestMain(m *testing.M) {
	var err error
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.Printf("local ip: %s\n", getLocalIPv4())
	// mysql
	_connectorGormDB, err = gorm.Open(mysql.Open(_connectorMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// redis
	opt, err := redis.ParseURL(_connectorRedisDsn)
	if err != nil {
		log.Fatalln(err)
	}
	_redisCli = redis.NewClient(opt)
	os.Exit(m.Run())
}

func getLocalIPv4() string {
	ips, _ := getLocalIPv4s()
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

func getLocalIPv4s() ([]string, error) {
	ips := make([]string, 0)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips, nil
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
