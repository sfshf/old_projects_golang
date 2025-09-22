package slark_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/ground/pkg/util"
	. "github.com/nextsurfer/slark/internal/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_slarkMysqlDsn = os.Getenv("SLARK_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_gormDB        *gorm.DB
	_slarkRedisDsn = os.Getenv("SLARK_REDIS_DNS")
	_redisCli      *redis.Client

	_kongDNS = "https://api.test.n1xt.net"

	// test account
	_testNickname    = "slark-e2e-test-1"
	_testEmail       = "slark@e2e-test-1.com"
	_testPassword    = "qwer1234"
	_testAccount     SlkUser
	_testApplication = "slark-e2e"
	_testDeviceID    = "NOID"
	_testSession     SlkSession
	_testCookie      *http.Cookie
)

func TestMain(m *testing.M) {
	var err error
	// mysql
	_gormDB, err = gorm.Open(mysql.Open(_slarkMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// redis
	opt, err := redis.ParseURL(_slarkRedisDsn)
	if err != nil {
		log.Fatalln(err)
	}
	_redisCli = redis.NewClient(opt)

	// check account and session records for testing
	h := sha3.NewKeccak256()
	_, err = h.Write([]byte(_testPassword))
	if err != nil {
		log.Fatalln(err)
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))
	if err := _gormDB.Table(TableNameSlkUser).
		Where(`email=? AND password_hash=? AND deleted_at=0`,
			_testEmail, testPasswordHash).
		First(&_testAccount).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testAccount.Nickname = _testNickname
			_testAccount.Email = _testEmail
			_testAccount.PasswordHash = testPasswordHash
			if err := _gormDB.Table(TableNameSlkUser).
				Create(&_testAccount).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	if err := _gormDB.Table(TableNameSlkSession).
		Where(`user_id = ? AND application = ? AND device_id = ? AND deleted_at = 0`,
			_testAccount.ID, _testApplication, _testDeviceID).
		First(&_testSession).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testSession.UserID = _testAccount.ID
			_testSession.SessionID = util.NewUUIDHexEncoding()
			_testSession.Application = _testApplication
			_testSession.DeviceID = _testDeviceID
			_testSession.LoginIP = getLocalIPv4()
			if err := _gormDB.Table(TableNameSlkSession).
				Create(&_testSession).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	_testCookie = &http.Cookie{
		HttpOnly: false,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    _testSession.SessionID,
	}

	// set ADJECTIVES_JSON_FILE and NOUNS_JSON_FILE
	os.Setenv("ADJECTIVES_JSON_FILE", "/app/configs/json/adjectives.json")
	os.Setenv("NOUNS_JSON_FILE", "/app/configs/json/nouns.json")
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

func getLoginEmailCaptchas() ([]struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}, error) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Email string `json:"email"`
				Code  string `json:"code"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/test/getLoginEmailCaptchas/v1", nil, nil, &respData, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || respData.Code != 0 {
		return nil, errors.New("bad request")
	}
	return respData.Data.List, nil
}

func getRegistrationEmailCaptchas() ([]struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}, error) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Email string `json:"email"`
				Code  string `json:"code"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/slark/test/getRegistrationEmailCaptchas/v1", nil, nil, &respData, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || respData.Code != 0 {
		return nil, errors.New("bad request")
	}
	return respData.Data.List, nil
}
