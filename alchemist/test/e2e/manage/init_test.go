package manage_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/ground/pkg/rpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_slarkMysqlDsn     = os.Getenv("SLARK_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_slarkGormDB       *gorm.DB
	_alchemistMysqlDsn = os.Getenv("ALCHEMIST_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_alchemistGormDB   *gorm.DB
	_alchemistRedisDsn = os.Getenv("ALCHEMIST_REDIS_DNS")
	_redisCli          *redis.Client

	_kongDNS = os.Getenv("ORACLE_GATEWAY_DNS")

	// test account
	_testNickname    = "alchemist-e2e-test-1"
	_testEmail       = "alchemist@e2e-test-1.com"
	_testPassword    = "qwer1234"
	_testAccount     SlkUser
	_testApplication = "alchemist-e2e"
	_testDeviceID    = "NOID"
	_testSession     SlkSession
	_testCookie      *http.Cookie

	// test account2
	_testNickname2 = "alchemist-e2e-test-2"
	_testEmail2    = "alchemist@e2e-test-2.com"
	_testAccount2  SlkUser
	_testSession2  SlkSession
	_testCookie2   *http.Cookie

	// some configs
	_DiscountOfferIDNewUser = "promo.discount.12m"
	_DiscountOfferID10M     = "promo.discount.10m"
	_DiscountOfferID8M      = "promo.discount.8m"
	_DiscountOfferID6M      = "promo.discount.6m"
	_DiscountOfferID4M      = "promo.discount.4m"
	_DiscountOfferID2M      = "promo.discount.2m"

	_alchemistApiKey = "DX7XHtLxlkzGYDtnmSACFqew"
)

func TestMain(m *testing.M) {
	var err error
	// mysql
	_slarkGormDB, err = gorm.Open(mysql.Open(_slarkMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	_alchemistGormDB, err = gorm.Open(mysql.Open(_alchemistMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// redis
	opt, err := redis.ParseURL(_alchemistRedisDsn)
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
			_testSession.SessionID = NewUUIDHexEncoding()
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
			_testSession2.SessionID = NewUUIDHexEncoding()
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

func NewUUIDHexEncoding() string {
	uuid := uuid.New()
	var buf [32]byte
	hex.Encode(buf[:], uuid[:])
	return strings.ToUpper(string(buf[:]))
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
