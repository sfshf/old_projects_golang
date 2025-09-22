package tester_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/tester/internal/pkg/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_kongDNS      = os.Getenv("ORACLE_GATEWAY_DNS")
	_testerApiKey = os.Getenv("TESTER_APIKEY")

	_slarkMysqlDsn = os.Getenv("SLARK_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_slarkGormDB   *gorm.DB
	_mongoDB       *mongo.Database

	// test account
	_testNickname    = "tester-e2e-test-1"
	_testEmail       = "tester@e2e-test-1.com"
	_testPassword    = "qwer1234"
	_testAccount     SlkUser
	_testApplication = "tester-e2e"
	_testDeviceID    = "NOID"
	_testSession     SlkSession
	_testCookie      *http.Cookie

	// test account2
	_testNickname2 = "tester-e2e-test-2"
	_testEmail2    = "tester@e2e-test-2.com"
	_testAccount2  SlkUser
	_testSession2  SlkSession
	_testCookie2   *http.Cookie

	// test account3
	_testNickname3 = "tester-e2e-test-3"
	_testEmail3    = "tester@e2e-test-3.com"
	_testAccount3  SlkUser
	_testSession3  SlkSession
	_testCookie3   *http.Cookie
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("CONSUL_HTTP_ADDR", "172.31.29.192:8500")

	var err error
	// mysql
	_slarkGormDB, err = gorm.Open(mysql.Open(_slarkMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// check account and session records for testing
	h := sha3.NewKeccak256()
	_, err = h.Write([]byte(_testPassword))
	if err != nil {
		log.Fatalln(err)
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))
	if err := _slarkGormDB.Table(TableNameSlkUser).
		Where(`nickname=? AND email=? AND password_hash=? AND deleted_at=0`,
			_testNickname, _testEmail, testPasswordHash).
		First(&_testAccount).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testAccount.Nickname = _testNickname
			_testAccount.Email = _testEmail
			_testAccount.PasswordHash = testPasswordHash
			if err := _slarkGormDB.Table(TableNameSlkUser).
				Create(&_testAccount).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	if err := _slarkGormDB.Table(TableNameSlkSession).
		Where(`user_id = ? AND application = ? AND device_id = ? AND deleted_at = 0`,
			_testAccount.ID, _testApplication, _testDeviceID).
		First(&_testSession).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testSession.UserID = _testAccount.ID
			_testSession.SessionID = uuid.NewUUIDHexEncoding()
			_testSession.Application = _testApplication
			_testSession.DeviceID = _testDeviceID
			_testSession.LoginIP = getLocalIPv4()
			if err := _slarkGormDB.Table(TableNameSlkSession).
				Create(&_testSession).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	_testCookie = &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    _testSession.SessionID,
	}
	log.Println("================>_testSession.SessionID:", _testSession.SessionID)
	// account2:
	if err := _slarkGormDB.Table(TableNameSlkUser).
		Where(`nickname=? AND email=? AND password_hash=? AND deleted_at=0`,
			_testNickname2, _testEmail2, testPasswordHash).
		First(&_testAccount2).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testAccount2.Nickname = _testNickname2
			_testAccount2.Email = _testEmail2
			_testAccount2.PasswordHash = testPasswordHash
			if err := _slarkGormDB.Table(TableNameSlkUser).
				Create(&_testAccount2).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	if err := _slarkGormDB.Table(TableNameSlkSession).
		Where(`user_id = ? AND application = ? AND device_id = ? AND deleted_at = 0`,
			_testAccount2.ID, _testApplication, _testDeviceID).
		First(&_testSession2).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testSession2.UserID = _testAccount2.ID
			_testSession2.SessionID = uuid.NewUUIDHexEncoding()
			_testSession2.Application = _testApplication
			_testSession2.DeviceID = _testDeviceID
			_testSession2.LoginIP = getLocalIPv4()
			if err := _slarkGormDB.Table(TableNameSlkSession).
				Create(&_testSession2).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	_testCookie2 = &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    _testSession2.SessionID,
	}
	log.Println("================>_testSession2.SessionID:", _testSession2.SessionID)
	// account3:
	if err := _slarkGormDB.Table(TableNameSlkUser).
		Where(`nickname=? AND email=? AND password_hash=? AND deleted_at=0`,
			_testNickname3, _testEmail3, testPasswordHash).
		First(&_testAccount3).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testAccount3.Nickname = _testNickname3
			_testAccount3.Email = _testEmail3
			_testAccount3.PasswordHash = testPasswordHash
			if err := _slarkGormDB.Table(TableNameSlkUser).
				Create(&_testAccount3).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	if err := _slarkGormDB.Table(TableNameSlkSession).
		Where(`user_id = ? AND application = ? AND device_id = ? AND deleted_at = 0`,
			_testAccount3.ID, _testApplication, _testDeviceID).
		First(&_testSession3).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testSession3.UserID = _testAccount3.ID
			_testSession3.SessionID = uuid.NewUUIDHexEncoding()
			_testSession3.Application = _testApplication
			_testSession3.DeviceID = _testDeviceID
			_testSession3.LoginIP = getLocalIPv4()
			if err := _slarkGormDB.Table(TableNameSlkSession).
				Create(&_testSession3).Error; err != nil {
				log.Fatalln(err)
			}
		}
	}
	_testCookie3 = &http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 10,
		Name:     rpc.DefaultCookieSessionKey,
		Value:    _testSession3.SessionID,
	}
	log.Println("================>_testSession3.SessionID:", _testSession3.SessionID)

	// mongo db
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Fatalln(err)
	}
	cliOpt := options.Client().ApplyURI(mongodbUri)
	mgoCli, err := mongo.Connect(cliOpt)
	if err != nil {
		log.Fatalln(err)
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Fatalln(err)
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "tester"
	}
	_mongoDB = mgoCli.Database(dbName)
	os.Exit(m.Run())
}

// slark models

const TableNameSlkUser = "slk_user"

type SlkUser struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:id" json:"id"` // id
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    int64     `gorm:"column:deleted_at;not null" json:"deleted_at"`
	Nickname     string    `gorm:"column:nickname;not null" json:"nickname"`
	PasswordHash string    `gorm:"column:password_hash;not null;comment:HASH" json:"password_hash"` // HASH
	Email        string    `gorm:"column:email;not null;comment:email," json:"email"`               // email,
	Phone        string    `gorm:"column:phone;not null;comment:," json:"phone"`                    // ,
}

const TableNameSlkSession = "slk_session"

type SlkSession struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:id" json:"id"` // id
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   int64     `gorm:"column:deleted_at;not null" json:"deleted_at"`
	Application string    `gorm:"column:application;not null" json:"application"`
	UserID      int64     `gorm:"column:user_id;not null;comment:id" json:"user_id"`               // id
	SessionID   string    `gorm:"column:session_id;not null;comment:session ID" json:"session_id"` // session ID
	DeviceID    string    `gorm:"column:device_id;not null;comment:id ID" json:"device_id"`        // id ID
	LoginIP     string    `gorm:"column:login_ip;not null;comment:ip, v4v6" json:"login_ip"`       // ip, v4v6
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
			log.Println(err)
			return nil, err
		}
		log.Printf("request data: %s\n", jsonData)
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
	if reqHeaderFunc != nil {
		reqHeaderFunc(req)
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
		log.Println(err)
		return nil, err
	}
	log.Printf("response data: %s\nresponse data size: %fM\n", data, float64(len(data))/1024/1024)
	if err := json.Unmarshal(data, respData); err != nil {
		log.Println(err)
		return resp, err
	}
	return resp, nil
}
