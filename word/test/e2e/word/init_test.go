package word_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/ground/pkg/rpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_slarkMysqlDsn = os.Getenv("SLARK_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_slarkGormDB   *gorm.DB
	_wordMysqlDsn  = os.Getenv("WORD_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_wordGormDB    *gorm.DB
	_wordRedisDsn  = os.Getenv("WORD_REDIS_DNS")
	_redisCli      *redis.Client

	_kongDNS = os.Getenv("ORACLE_GATEWAY_DNS")

	// test account
	_testNickname    = "word-e2e-test-1"
	_testEmail       = "word@e2e-test-1.com"
	_testPassword    = "qwer1234"
	_testAccount     SlkUser
	_testApplication = "word-e2e"
	_testDeviceID    = "NOID"
	_testSession     SlkSession
	_testCookie      *http.Cookie

	_progressBackupBucketName = "n1xt-backup-test"
	_s3Client                 *s3.Client
)

func TestMain(m *testing.M) {
	var err error
	// mysql
	_slarkGormDB, err = gorm.Open(mysql.Open(_slarkMysqlDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_wordGormDB, err = gorm.Open(mysql.Open(_wordMysqlDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// redis
	opt, err := redis.ParseURL(_wordRedisDsn)
	if err != nil {
		panic(err)
	}
	_redisCli = redis.NewClient(opt)

	// check account and session records for testing
	h := sha3.NewKeccak256()
	_, err = h.Write([]byte(_testPassword))
	if err != nil {
		panic(err)
	}
	testPasswordHash := hex.EncodeToString(h.Sum(nil))
	if err := _slarkGormDB.Table(TableNameSlkUser).
		Where(`nickname=? AND email=? AND password_hash=? AND deleted_at=0`,
			_testNickname, _testEmail, testPasswordHash).
		First(&_testAccount).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		} else {
			_testAccount.Nickname = _testNickname
			_testAccount.Email = _testEmail
			_testAccount.PasswordHash = testPasswordHash
			if err := _slarkGormDB.Table(TableNameSlkUser).
				Create(&_testAccount).Error; err != nil {
				panic(err)
			}
		}
	}
	if err := _slarkGormDB.Table(TableNameSlkSession).
		Where(`user_id = ? AND application = ? AND device_id = ?  AND deleted_at = 0`,
			_testAccount.ID, _testApplication, _testDeviceID).
		First(&_testSession).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		} else {
			_testSession.UserID = _testAccount.ID
			_testSession.SessionID = NewUUIDHexEncoding()
			_testSession.Application = _testApplication
			_testSession.DeviceID = _testDeviceID
			_testSession.LoginIP = getLocalIPv4()
			if err := _slarkGormDB.Table(TableNameSlkSession).
				Create(&_testSession).Error; err != nil {
				panic(err)
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

	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("STUDY_BACKUP_KEY", "860a5b914f67c287210e01a9eac15feb")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithLogger(logger{}))
	if err != nil {
		panic(fmt.Sprintf("aws sdk LoadDefaultConfig failed: %s", err))
	}
	// Create an Amazon S3 service client
	_s3Client = s3.NewFromConfig(cfg)
	os.Exit(m.Run())
}

type logger struct{}

func (logger) Logf(classification logging.Classification, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		log.Println("AWS Warning: ", msg)
	} else if classification == logging.Debug {
		log.Println("AWS Debug: ", msg)
	}
}

func md5Sum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func getHashedPath(version int32, userID, timestamp int64, data []byte) string {
	return strings.ToLower(fmt.Sprintf("%d/%d/%d.%x", version, userID, timestamp, md5Sum(data)))
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
	log.Printf("response data: %s\n", data)
	if err := json.Unmarshal(data, respData); err != nil {
		log.Println(err)
		return resp, err
	}
	return resp, nil
}
