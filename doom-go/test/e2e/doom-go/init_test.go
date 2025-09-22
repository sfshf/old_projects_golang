package doom_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	ecies "github.com/ecies/go/v2"
	"github.com/google/uuid"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
	"github.com/nextsurfer/ground/pkg/localize"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_kongDNS   = "https://api.test.n1xt.net"
	_publicKey *ecies.PublicKey

	_slarkMysqlDsn = os.Getenv("SLARK_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_slarkGormDB   *gorm.DB

	// test account
	_testEmail       = "doom@e2e-test.com"
	_testPassword    = "qwer1234"
	_testAccount     SlkUser
	_testApplication = "doom-e2e"
	_testDeviceID    = "NOID"
	_testSession     SlkSession
	_testCookie      *http.Cookie

	_connectorApiKey  = os.Getenv("CONNECTOR_API_KEY")
	_connectorKeyID   = os.Getenv("CONNECTOR_KEY_ID")
	_localizerManager = localize.NewManager()

	_mongoDB *mongo.Database
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("CONSUL_HTTP_ADDR", "172.31.29.192:8500")

	if err := CheckKeyExisting(_kongDNS, _connectorApiKey, _connectorKeyID); err != nil {
		log.Fatalln(err)
	}

	hexPublicKey, err := GetPublicKey(_kongDNS, _connectorApiKey, _connectorKeyID)
	if err != nil {
		log.Fatalln(err)
	}
	pubKeyBytes, err := hex.DecodeString(hexPublicKey)
	if err != nil {
		log.Fatalln(err)
	}
	_publicKey, err = ecies.NewPublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		log.Fatalln(err)
	}

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
		Where(`email = ? AND password_hash = ? AND deleted_at = 0`,
			_testEmail, testPasswordHash).
		First(&_testAccount).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalln(err)
		} else {
			_testAccount.Email = _testEmail
			_testAccount.Nickname = "doom-go-e2e"
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
	// mongo db
	mongodbUri := "mongodb+srv://sheldon:obZZHKYcMhrPIavE@test1.6sj0f.mongodb.net/?retryWrites=true&w=majority&appName=Test1"
	uri, err := url.Parse(mongodbUri)
	if err != nil {
		log.Fatalln(err)
	}
	cliOpt := options.Client().ApplyURI(mongodbUri)
	mgoCli, err := mongo.Connect(ctx, cliOpt)
	if err != nil {
		log.Fatalln(err)
	}
	if err := mgoCli.Ping(ctx, nil); err != nil {
		log.Fatalln(err)
	}
	dbName := uri.Path[1:]
	if dbName == "" {
		dbName = "doom"
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

// connector models

const TableNameRelationAppDatum = "relation_app_data"

// RelationAppDatum mapped from table <relation_app_data>
type RelationAppDatum struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt int64     `gorm:"column:deleted_at;not null;comment:Coding style" json:"deleted_at"` // Coding style
	App       string    `gorm:"column:app;not null" json:"app"`
	KeyID     string    `gorm:"column:key_id;not null" json:"key_id"`
	DataID    string    `gorm:"column:data_id;not null" json:"data_id"`
}

func NewUUIDHexEncoding() string {
	uuid := uuid.New()
	var buf [32]byte
	hex.Encode(buf[:], uuid[:])
	return strings.ToUpper(string(buf[:]))
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

func CheckKeyExisting(kongDNS, apiKey, keyID string) error {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: apiKey,
		KeyID:  keyID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Existing bool `json:"existing"`
		} `json:"data"`
	}{}
	resp, err := simplehttp.PostJsonRequest(kongDNS+"/riki/checkKeyExisting/v1", &reqData, nil, &respData)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("check key existing: http code not equal to 200")
	}
	if respData.Code != 0 {
		return fmt.Errorf("check key existing: %v", respData)
	}
	if !respData.Data.Existing {
		return errors.New("key id not exist")
	}
	return nil
}

func GetPublicKey(kongDNS, apiKey, keyID string) (string, error) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: apiKey,
		KeyID:  keyID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	// send request
	resp, err := simplehttp.PostJsonRequest(kongDNS+"/riki/getPublicKey/v1", &reqData, nil, &respData)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("get public key: http code not equal to 200")
	}
	if respData.Code != 0 {
		return "", fmt.Errorf("get public key: %v", respData)
	}
	if respData.Data.PublicKey == "" {
		return "", fmt.Errorf("publick key is empty: %v", respData)
	}
	return respData.Data.PublicKey, nil
}

const (
	PasswordLength = 16
)

func Random(length int) []byte {
	if length <= 0 {
		length = 6
	}
	rands, err := randomBytesMod(length, 36)
	if err != nil {
		return nil
	}
	var buf bytes.Buffer
	for _, rand := range rands {
		if rand < 10 {
			buf.WriteRune(rune(rand + 48))
		} else {
			buf.WriteRune(rune(rand + 87))
		}
	}
	return buf.Bytes()
}

func randomBytesMod(length int, mod byte) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be greater than zero")
	}
	if mod <= 0 {
		return nil, errors.New("captcha: bad mod argument for randomBytesMod")
	}
	maxrb := 255 - byte(256%int(mod))
	b := make([]byte, length)
	i := 0
	for {
		r, err := randomBytes(length + (length / 4))
		if err != nil {
			return nil, err
		}
		for _, c := range r {
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = c % mod
			i++
			if i == length {
				return b, nil
			}
		}
	}
}

func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("captcha: error reading random source: %v", err)
	}
	return b, nil
}

func Keccak256(src []byte) ([]byte, error) {
	h := sha3.NewKeccak256()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func Keccak256Hex(src []byte) ([]byte, error) {
	sum, err := Keccak256(src)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return dst, nil
}
