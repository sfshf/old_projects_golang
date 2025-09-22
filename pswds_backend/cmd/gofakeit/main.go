package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	ecies "github.com/ecies/go/v2"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/common/simplecrypto"
	"github.com/nextsurfer/pswds_backend/internal/common/simplehttp"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// -------------------------------------------------------- slark --------------------------------------------------------

var (
	_slarkGormDB     *gorm.DB
	_initSlarkEmails = []string{
		"test001@gmail.com", // first family -- creator
		"test002@gmail.com", // first family -- not admin
		"test003@gmail.com", // no family -- has first family invitation
		"test004@gmail.com", // no family -- no family invitation
		"test005@gmail.com", // second family -- creator
		"test006@gmail.com", // second family -- admin
		"test007@gmail.com", // second family -- not admin
		"test008@gmail.com", // no family -- has second family invitation
		"test009@gmail.com", // no family -- no family invitation
	}
	_initSlarkAccounts              map[string]*SlkUser // email -> account
	_initSlarkPassword              = "111111"
	_initSlarkPasswordHash          string
	_initSlarkSecondaryPassword     = "111111"
	_initSlarkSecondaryPasswordHash string
)

// slark tables

const TableNameSlkUser = "slk_user"

type SlkUser struct {
	ID                    int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:id" json:"id"` // id
	CreatedAt             time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt             int64     `gorm:"column:deleted_at;not null" json:"deleted_at"`
	Nickname              string    `gorm:"column:nickname;not null" json:"nickname"`
	PasswordHash          string    `gorm:"column:password_hash;comment:password hash" json:"password_hash"`                               // password hash
	SecondaryPasswordHash string    `gorm:"column:secondary_password_hash;comment:secondary password hash" json:"secondary_password_hash"` // secondary password hash
	Email                 string    `gorm:"column:email;comment:email address" json:"email"`                                               // email address
	Phone                 string    `gorm:"column:phone;comment:phone number" json:"phone"`                                                // phone number
}

// -------------------------------------------------------- pswds --------------------------------------------------------

var (
	_pswdsGormDB     *gorm.DB
	_pswdsDaoManager *dao.Manager

	needToDeletedUserIDs             []int64
	_initUserPrivKey                 *ecies.PrivateKey
	_initFamilyKey                   []byte
	_initEncryptedFamilyKey          string
	_initSecurityQuestions           string
	_initSecurityQuestionsCiphertext string
	_firstFamilyMembers              []int64
	_secondFamilyMembers             []int64

	_initTrustedContactPassword = "111111"
)

func init() {
	// 1. log settings
	log.SetFlags(log.LstdFlags | log.Llongfile)
	// gofakeit seed
	gofakeit.Seed(time.Now().UnixNano())
	// slark password hash
	data, err := simplecrypto.Keccak256Hex([]byte(_initSlarkPassword))
	if err != nil {
		log.Fatalln(err)
	}
	_initSlarkPasswordHash = string(data)
	data, err = simplecrypto.Keccak256Hex([]byte(_initSlarkSecondaryPassword))
	if err != nil {
		log.Fatalln(err)
	}
	_initSlarkSecondaryPasswordHash = string(data)
	// 2. mysql connections
	slarkMysqlDns := os.Getenv("SLARK_MYSQL_DNS")
	if slarkMysqlDns == "" {
		log.Fatalln("slark msyql dns is empty")
	}
	_slarkGormDB, err = gorm.Open(mysql.Open(slarkMysqlDns), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	pswdsMysqlDns := os.Getenv("PSWDS_MYSQL_DNS")
	if pswdsMysqlDns == "" {
		log.Fatalln("pswds msyql dns is empty")
	}
	_pswdsGormDB, err = gorm.Open(mysql.Open(pswdsMysqlDns), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	_pswdsDaoManager = dao.ManagerWithDB(_pswdsGormDB)
	// 3. slark accounts cache
	_initSlarkAccounts = make(map[string]*SlkUser)
	// 4.pswds preludes
	// 4-1. user private key
	_initUserPrivKey, err = UserPrivateKey(_initSlarkSecondaryPassword)
	if err != nil {
		log.Fatalln(err)
	}
	// 4-2. family key
	_initFamilyKey, err = FamilyKey(_initSlarkSecondaryPassword)
	if err != nil {
		log.Fatalln(err)
	}
	initEncryptedFamilyKeyBytes, err := EncryptedFamilyKey(_initFamilyKey, _initUserPrivKey)
	if err != nil {
		log.Fatalln(err)
	}
	_initEncryptedFamilyKey = hex.EncodeToString(initEncryptedFamilyKeyBytes)
	// 4-3. security questions
	questions := struct {
		Question1 string `json:"question1"`
		Question2 string `json:"question2"`
		Question3 string `json:"question3"`
	}{
		Question1: "question1",
		Question2: "question2",
		Question3: "question3",
	}
	questionsJson, err := json.Marshal(questions)
	if err != nil {
		log.Fatalln(err)
	}
	_initSecurityQuestions = string(questionsJson)
	answers := struct {
		Answer1 string `json:"answer1"`
		Answer2 string `json:"answer2"`
		Answer3 string `json:"answer3"`
	}{
		Answer1: "answer1",
		Answer2: "answer2",
		Answer3: "answer3",
	}
	answersHash, err := simplecrypto.Keccak256Hex([]byte(answers.Answer1 + answers.Answer2 + answers.Answer3))
	if err != nil {
		log.Fatalln(err)
	}
	hashOfHash, err := simplecrypto.Keccak256Hex(answersHash)
	if err != nil {
		log.Fatalln(err)
	}
	reqData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: "fJVfDpPciWfym6KK6dblaEmw", // pswd_pswds
		KeyID:  "pswds-001",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"` // hex string
		} `json:"data"`
	}{}
	resp, err := simplehttp.PostJsonRequest("https://api.test.n1xt.net/riki/getPublicKey/v1", &reqData, nil, &respData)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != http.StatusOK || respData.Data.PublicKey == "" {
		log.Fatalln("fetch pswds-001 public key fail")
	}
	nonce, aead, err := simplecrypto.NewNonceAndX(answersHash[:32])
	if err != nil {
		log.Fatalln(err)
	}
	encryptedUnlockPassword := string(simplecrypto.EncryptByX([]byte(_initSlarkSecondaryPassword), aead, nonce))
	plainObj := struct {
		Question1         string `json:"question1"`
		Question2         string `json:"question2"`
		Question3         string `json:"question3"`
		EncryptedPassword string `json:"encryptedPassword"`
		AnswerHash        string `json:"answerHash"`
		HashOfHash        string `json:"hashOfHash"`
		Nonce             string `json:"nonce"`
	}{
		Question1:         questions.Question1,
		Question2:         questions.Question2,
		Question3:         questions.Question3,
		EncryptedPassword: encryptedUnlockPassword,
		AnswerHash:        string(answersHash),
		HashOfHash:        string(hashOfHash),
		Nonce:             base64.StdEncoding.EncodeToString(nonce),
	}
	plainBytes, err := json.Marshal(plainObj)
	if err != nil {
		log.Fatalln(err)
	}
	b64Plaintext := base64.StdEncoding.EncodeToString(plainBytes)
	pubKey, err := ecies.NewPublicKeyFromHex(respData.Data.PublicKey)
	if err != nil {
		log.Fatalln(err)
	}
	initSecurityQuestionsCiphertext, err := ecies.Encrypt(pubKey, []byte(b64Plaintext))
	if err != nil {
		log.Fatalln(err)
	}
	_initSecurityQuestionsCiphertext = base64.StdEncoding.EncodeToString(initSecurityQuestionsCiphertext)
}

func xorStr(secretHex string, origin string) string {
	if origin == "" {
		return ""
	}
	secret, _ := hex.DecodeString(secretHex)
	return hex.EncodeToString(xor(secret, []byte(origin)))
}

func xor(secret []byte, origin []byte) []byte {
	var index int
	handled := make([]byte, len(origin))
	for i := range len(origin) {
		val, next := next(secret, index)
		handled[i] = origin[i] ^ val
		index = next
	}
	return handled
}

func next(secret []byte, index int) (byte, int) {
	if index == len(secret) {
		index = 0
	}
	return secret[index], index + 1
}

type XoredPassword struct {
	DataID         string `json:"dataID" fake:"-"`
	CreatedAt      string `json:"createdAt" fake:"-"`
	UpdatedAt      string `json:"updatedAt" fake:"-"`
	UserID         string `json:"userID" fake:"-"`
	Title          string `json:"title" fake:"-"`
	Website        string `json:"website" fake:"-"`
	Username       string `json:"username" fake:"-"`
	Password       string `json:"password" fake:"-"`
	Notes          string `json:"notes" fake:"-"`
	Others         string `json:"others" fake:"-"`
	UsedAt         string `json:"usedAt" fake:"-"`
	UsedCount      string `json:"usedCount" fake:"-"`
	IconBgColor    string `json:"iconBgColor" fake:"-"` // 0, 1, 2, 3
	SharedAt       string `json:"sharedAt" fake:"-"`
	SharedToAll    string `json:"sharedToAll" fake:"-"` // 1
	SharingMembers string `json:"sharingMembers" fake:"-"`
}

type Password struct {
	DataID         string `json:"dataID" fake:"{uuid}"`
	CreatedAt      int64  `json:"createdAt" fake:"-"`
	UpdatedAt      int64  `json:"updatedAt" fake:"-"`
	UserID         int64  `json:"userID" fake:"-"`
	Title          string `json:"title" fake:"{sentence:5}"`
	Website        string `json:"website" fake:"-"`
	Username       string `json:"username" fake:"{firstname}"`
	Password       string `json:"password" fake:"-"`
	Notes          string `json:"notes" fake:"{sentence:30}"`
	Others         string `json:"others" fake:"-"`
	UsedAt         int64  `json:"usedAt" fake:"-"`
	UsedCount      int64  `json:"usedCount" fake:"-"`
	IconBgColor    int    `json:"iconBgColor" fake:"{randomint:[0,1,2,3]}"` // 0, 1, 2, 3
	SharedAt       int64  `json:"sharedAt" fake:"-"`
	SharedToAll    int    `json:"sharedToAll" fake:"-"` // 1
	SharingMembers string `json:"sharingMembers" fake:"-"`
}

func xorPassword(secretHex string, entity Password) XoredPassword {
	var result XoredPassword
	result.DataID = xorStr(secretHex, entity.DataID)
	result.CreatedAt = xorStr(secretHex, strconv.Itoa(int(entity.CreatedAt)))
	result.UpdatedAt = xorStr(secretHex, strconv.Itoa(int(entity.UpdatedAt)))
	result.UserID = xorStr(secretHex, strconv.Itoa(int(entity.UserID)))
	result.Title = xorStr(secretHex, entity.Title)
	result.Website = xorStr(secretHex, entity.Website)
	result.Username = xorStr(secretHex, entity.Username)
	result.Password = xorStr(secretHex, entity.Password)
	result.Notes = xorStr(secretHex, entity.Notes)
	result.Others = xorStr(secretHex, entity.Others)
	result.UsedAt = xorStr(secretHex, strconv.Itoa(int(entity.UsedAt)))
	result.UsedCount = xorStr(secretHex, strconv.Itoa(int(entity.UsedCount)))
	result.IconBgColor = xorStr(secretHex, strconv.Itoa(int(entity.IconBgColor)))
	result.SharedAt = xorStr(secretHex, strconv.Itoa(int(entity.SharedAt)))
	result.SharedToAll = xorStr(secretHex, strconv.Itoa(int(entity.SharedToAll)))
	result.SharingMembers = xorStr(secretHex, entity.SharingMembers)
	return result
}

func xorPasswords(secretHex string, entities []Password) []XoredPassword {
	var result []XoredPassword
	for _, item := range entities {
		result = append(result, xorPassword(secretHex, item))
	}
	return result
}

type Other struct {
	Type  string `json:"type" fake:"{randomstring:[text,url,password,one-time password,date,pin]}"`
	Key   string `json:"key" fake:"{booktitle}"`
	Value string `json:"value" fake:"-"`
}

func RandomPassword() string {
	length := 14
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var newPassword string
	// must number
	var mustNumber string
	numbers := "0123456789"
	charset += numbers
	index := mathrand.Intn(len(numbers))
	mustNumber = numbers[index : index+1]
	// must symbol
	var mustSymbol string
	symbols := "_!@#$%^&*()"
	charset += symbols
	index = mathrand.Intn(len(symbols))
	mustSymbol = symbols[index : index+1]
	mustNumberIndex := mathrand.Intn(length)
	mustSymbolIndex := mathrand.Intn(length)
	if mustSymbolIndex == mustNumberIndex {
		mustSymbolIndex += 1
	}
	for i := 0; i < length; i++ {
		if i == mustNumberIndex {
			newPassword += mustNumber
		} else if i == mustSymbolIndex {
			newPassword += mustSymbol
		} else {
			index := mathrand.Intn(len(charset))
			newPassword += charset[index : index+1]
		}
	}
	return newPassword
}

func RandomPIN(length int) string {
	if length == 0 {
		length := mathrand.Intn(12)
		if length < 4 {
			length = 4
		}
	}
	numbers := "0123456789"
	var pin string
	for i := 0; i < length; i++ {
		index := mathrand.Intn(len(numbers))
		pin += numbers[index : index+1]
	}
	return pin
}

// otpauth://totp/GitHub:evilsophietheking?secret=ECX3VKG4XX2D5V55&issuer=GitHub
func RandomOthers() string {
	length := mathrand.Intn(10)
	if length == 0 {
		length = 3
	}
	var others []Other
	for range length {
		var one Other
		if err := gofakeit.Struct(&one); err != nil {
			log.Fatalln(err)
		}
		switch one.Type {
		case "text":
			one.Value = gofakeit.Sentence(30)
		case "url":
			one.Value = gofakeit.URL()
		case "password":
			one.Value = RandomPassword()
		case "one-time password":
			one.Value = "otpauth://totp/GitHub:evilsophietheking?secret=ECX3VKG4XX2D5V55&issuer=GitHub"
		case "date":
			one.Value = gofakeit.Date().Format("2006-01-02")
		case "pin":
			one.Value = RandomPIN(0)
		}
		others = append(others, one)
	}
	data, err := json.Marshal(others)
	if err != nil {
		log.Fatalln(err)
	}
	return string(data)
}

func EncryptByUnlockPassword(data []byte) string {
	encryptKey, err := simplecrypto.Keccak256([]byte(_initSlarkSecondaryPassword + "9C9B913EB1B6254F4737CE947EFD16F16E916F"))
	if err != nil {
		log.Fatalln(err)
	}
	_, aead, err := simplecrypto.NewNonceAndX(encryptKey[:32])
	if err != nil {
		log.Fatalln(err)
	}
	return string(simplecrypto.EncryptByX(data, aead, simplecrypto.NonceZeroX()))
}

func NewPassword(userID int64, familyMembers []int64) *Password {
	ts := time.Now()
	var one Password
	if err := gofakeit.Struct(&one); err != nil {
		log.Fatalln(err)
	}
	one.CreatedAt = ts.Unix()
	one.UpdatedAt = ts.Unix()
	one.UserID = userID
	// random website
	var website string
	switch mathrand.Intn(3) {
	case 0:
		website = ""
	case 1:
		website = "https://www.baidu.com"
	case 2:
		website = "https://www.alipay.com"
	}
	one.Website = website
	// random password
	one.Password = RandomPassword()
	// random others
	one.Others = RandomOthers()
	// random family shares
	if len(familyMembers) > 0 {
		if mathrand.Intn(2) == 1 {
			one.SharedAt = ts.Unix()
			one.SharedToAll = mathrand.Intn(2)
			var sharingMembers []int64
			for _, member := range familyMembers {
				if mathrand.Intn(2) == 1 {
					sharingMembers = append(sharingMembers, member)
				}
			}
			if one.SharedToAll != 1 && len(sharingMembers) > 0 {
				data, err := json.Marshal(sharingMembers)
				if err != nil {
					log.Fatalln(err)
				}
				one.SharingMembers = string(data)
			}
		}
	}
	return &one
}

func InsertPasswordsAndSharedPasswordsByUserID(ctx context.Context, userID int64, familyID string, familyMembers []int64, length int) {
	otherMembers := excludeMyself(userID, familyMembers)
	var passwordRecords []*PasswordRecord
	var familySharedRecords []*FamilySharedRecord
	for range length {
		password := NewPassword(userID, otherMembers)
		data, err := json.Marshal(xorPassword(_initSlarkSecondaryPasswordHash, *password))
		if err != nil {
			log.Fatalln(err)
		}
		passwordRecords = append(passwordRecords, &PasswordRecord{
			DataID:  password.DataID,
			UserID:  userID,
			Content: EncryptByUnlockPassword(data),
			Version: 1,
		})
		if password.SharedAt > 0 {
			content, err := EncryptByFamilyKey(_initFamilyKey, data)
			if err != nil {
				log.Fatalln(err)
			}
			familySharedRecords = append(familySharedRecords, &FamilySharedRecord{
				DataID:         password.DataID,
				FamilyID:       familyID,
				SharedBy:       password.UserID,
				Type:           "password",
				Content:        string(content),
				SharedToAll:    int32(password.SharedToAll),
				SharingMembers: password.SharingMembers,
				Version:        1,
			})
		}
	}
	if err := _pswdsDaoManager.PasswordRecordDAO.Create(ctx, passwordRecords); err != nil {
		log.Fatalln(err)
	}
	if len(familySharedRecords) > 0 {
		if err := _pswdsDaoManager.FamilySharedRecordDAO.Create(ctx, familySharedRecords); err != nil {
			log.Fatalln(err)
		}
	}
}

func excludeMyself(myself int64, familyMembers []int64) []int64 {
	if len(familyMembers) == 0 {
		return nil
	}
	var members []int64
	for _, member := range familyMembers {
		if member != myself {
			members = append(members, member)
		}
	}
	return members
}

type XoredRecord struct {
	DataID      string `json:"dataID" fake:"-"`
	CreatedAt   string `json:"createdAt" fake:"-"`
	UpdatedAt   string `json:"updatedAt" fake:"-"`
	UserID      string `json:"userID" fake:"-"`
	RecordType  string `json:"recordType" fake:"-"`
	Title       string `json:"title" fake:"-"`
	IconBgColor string `json:"iconBgColor" fake:"-"` // 0, 1, 2, 3
	UsedAt      string `json:"usedAt" fake:"-"`
	UsedCount   string `json:"usedCount" fake:"-"`
	// mixed fields
	Phone          string `json:"phone" fake:"-"`
	Type           string `json:"type" fake:"-"`
	Number         string `json:"number" fake:"-"`
	Address        string `json:"address" fake:"-"`
	FullName       string `json:"fullName" fake:"-"`
	BirthDate      string `json:"birthDate" fake:"-" format:"2006-01-02"`
	Gender         string `json:"gender" fake:"-"`
	Pin            string `json:"pin" fake:"-"`
	ExpiryDate     string `json:"expiryDate" fake:"-" format:"2006-01-02"`
	Others         string `json:"others" fake:"-"`
	SharedAt       string `json:"sharedAt" fake:"-"`
	SharedToAll    string `json:"sharedToAll" fake:"-"` // 1
	SharingMembers string `json:"sharingMembers" fake:"-"`
	// identity fields
	FirstName            string `json:"firstName" fake:"-"`
	LastName             string `json:"lastName" fake:"-"`
	Job                  string `json:"job" fake:"-"`
	SocialSecurityNumber string `json:"socialSecurityNumber" fake:"-"`
	IdNumber             string `json:"idNumber" fake:"-"`
	// credit card fields
	CardholderName     string `json:"cardholderName" fake:"-"`
	VerificationNumber string `json:"verificationNumber" fake:"-"`
	ValidFrom          string `json:"validFrom" fake:"-"`
	IssuingBank        string `json:"issuingBank" fake:"-"`
	// bank account fields
	BankName      string `json:"bankName" fake:"-"`
	NameOnAccount string `json:"nameOnAccount" fake:"-"`
	RoutingNumber string `json:"routingNumber" fake:"-"`
	Branch        string `json:"branch" fake:"-"`
	AccountNumber string `json:"accountNumber" fake:"-"`
	Swift         string `json:"swift" fake:"-"`
	// driver license fields
	Height       string `json:"height" fake:"-"`
	LicenseClass string `json:"licenseClass" fake:"-"`
	State        string `json:"state" fake:"-"`
	Country      string `json:"country" fake:"-"`
	// passport fields
	IssuingCountry   string `json:"issuingCountry" fake:"-"`
	Nationality      string `json:"nationality" fake:"-"`
	IssuingAuthority string `json:"issuingAuthority" fake:"-"`
	BirthPlace       string `json:"birthPlace" fake:"-"`
	IssuedOn         string `json:"issuedOn" fake:"-"`
}

type Record struct {
	DataID      string `json:"dataID" fake:"{uuid}"`
	CreatedAt   int64  `json:"createdAt" fake:"-"`
	UpdatedAt   int64  `json:"updatedAt" fake:"-"`
	UserID      int64  `json:"userID" fake:"-"`
	RecordType  string `json:"recordType" fake:"{randomstring:[identity,credit card,bank account,driver license,passport]}"`
	Title       string `json:"title" fake:"{sentence:5}"`
	IconBgColor int    `json:"iconBgColor" fake:"{randomint:[0,1,2,3]}"` // 0, 1, 2, 3
	UsedAt      int64  `json:"usedAt" fake:"-"`
	UsedCount   int64  `json:"usedCount" fake:"-"`
	// mixed fields
	Phone          string `json:"phone" fake:"{phoneformatted}"`
	Type           string `json:"type" fake:"{sentence:3}"`
	Number         string `json:"number" fake:"{creditcardnumber}"`
	Address        string `json:"address" fake:"{address}"`
	FullName       string `json:"fullName" fake:"-"`
	BirthDate      string `json:"birthDate" fake:"{year}-{month}-{day}" format:"2006-01-02"`
	Gender         string `json:"gender" fake:"{gender}"`
	Pin            string `json:"pin" fake:"-"`
	ExpiryDate     string `json:"expiryDate" fake:"{year}-{month}-{day}" format:"2006-01-02"`
	Others         string `json:"others" fake:"-"`
	SharedAt       int64  `json:"sharedAt" fake:"-"`
	SharedToAll    int    `json:"sharedToAll" fake:"-"` // 1
	SharingMembers string `json:"sharingMembers" fake:"-"`
	// identity fields
	FirstName            string `json:"firstName" fake:"-"`
	LastName             string `json:"lastName" fake:"-"`
	Job                  string `json:"job" fake:"-"`
	SocialSecurityNumber string `json:"socialSecurityNumber" fake:"-"`
	IdNumber             string `json:"idNumber" fake:"-"`
	// credit card fields
	CardholderName     string `json:"cardholderName" fake:"-"`
	VerificationNumber string `json:"verificationNumber" fake:"-"`
	ValidFrom          string `json:"validFrom" fake:"-"`
	IssuingBank        string `json:"issuingBank" fake:"-"`
	// bank account fields
	BankName      string `json:"bankName" fake:"-"`
	NameOnAccount string `json:"nameOnAccount" fake:"-"`
	RoutingNumber string `json:"routingNumber" fake:"-"`
	Branch        string `json:"branch" fake:"-"`
	AccountNumber string `json:"accountNumber" fake:"-"`
	Swift         string `json:"swift" fake:"-"`
	// driver license fields
	Height       string `json:"height" fake:"-"`
	LicenseClass string `json:"licenseClass" fake:"-"`
	State        string `json:"state" fake:"-"`
	Country      string `json:"country" fake:"-"`
	// passport fields
	IssuingCountry   string `json:"issuingCountry" fake:"-"`
	Nationality      string `json:"nationality" fake:"-"`
	IssuingAuthority string `json:"issuingAuthority" fake:"-"`
	BirthPlace       string `json:"birthPlace" fake:"-"`
	IssuedOn         string `json:"issuedOn" fake:"-"`
}

func xorRecord(secretHex string, entity Record) XoredRecord {
	var result XoredRecord
	result.DataID = xorStr(secretHex, entity.DataID)
	result.CreatedAt = xorStr(secretHex, strconv.Itoa(int(entity.CreatedAt)))
	result.UpdatedAt = xorStr(secretHex, strconv.Itoa(int(entity.UpdatedAt)))
	result.UserID = xorStr(secretHex, strconv.Itoa(int(entity.UserID)))
	result.RecordType = xorStr(secretHex, entity.RecordType)
	result.Title = xorStr(secretHex, entity.Title)
	result.UsedAt = xorStr(secretHex, strconv.Itoa(int(entity.UsedAt)))
	result.UsedCount = xorStr(secretHex, strconv.Itoa(int(entity.UsedCount)))
	result.Others = xorStr(secretHex, entity.Others)
	result.IconBgColor = xorStr(secretHex, strconv.Itoa(int(entity.IconBgColor)))
	result.SharedAt = xorStr(secretHex, strconv.Itoa(int(entity.SharedAt)))
	result.SharedToAll = xorStr(secretHex, strconv.Itoa(int(entity.SharedToAll)))
	result.SharingMembers = xorStr(secretHex, entity.SharingMembers)
	result.Phone = xorStr(secretHex, entity.Phone)
	result.Type = xorStr(secretHex, entity.Type)
	result.Number = xorStr(secretHex, entity.Number)
	result.Address = xorStr(secretHex, entity.Address)
	result.FullName = xorStr(secretHex, entity.FullName)
	result.BirthDate = xorStr(secretHex, entity.BirthDate)
	result.Gender = xorStr(secretHex, entity.Gender)
	result.Pin = xorStr(secretHex, entity.Pin)
	result.ExpiryDate = xorStr(secretHex, entity.ExpiryDate)
	result.FirstName = xorStr(secretHex, entity.FirstName)
	result.LastName = xorStr(secretHex, entity.LastName)
	result.Job = xorStr(secretHex, entity.Job)
	result.SocialSecurityNumber = xorStr(secretHex, entity.SocialSecurityNumber)
	result.IdNumber = xorStr(secretHex, entity.IdNumber)
	result.CardholderName = xorStr(secretHex, entity.CardholderName)
	result.VerificationNumber = xorStr(secretHex, entity.VerificationNumber)
	result.ValidFrom = xorStr(secretHex, entity.ValidFrom)
	result.IssuingBank = xorStr(secretHex, entity.IssuingBank)
	result.BankName = xorStr(secretHex, entity.BankName)
	result.NameOnAccount = xorStr(secretHex, entity.NameOnAccount)
	result.RoutingNumber = xorStr(secretHex, entity.RoutingNumber)
	result.Branch = xorStr(secretHex, entity.Branch)
	result.AccountNumber = xorStr(secretHex, entity.AccountNumber)
	result.Swift = xorStr(secretHex, entity.Swift)
	result.Height = xorStr(secretHex, entity.Height)
	result.LicenseClass = xorStr(secretHex, entity.LicenseClass)
	result.State = xorStr(secretHex, entity.State)
	result.Country = xorStr(secretHex, entity.Country)
	result.IssuingCountry = xorStr(secretHex, entity.IssuingCountry)
	result.Nationality = xorStr(secretHex, entity.Nationality)
	result.IssuingAuthority = xorStr(secretHex, entity.IssuingAuthority)
	result.BirthPlace = xorStr(secretHex, entity.BirthPlace)
	result.IssuedOn = xorStr(secretHex, entity.IssuedOn)
	return result
}

func xorRecords(secretHex string, entities []Record) []XoredRecord {
	var result []XoredRecord
	for _, item := range entities {
		result = append(result, xorRecord(secretHex, item))
	}
	return result
}

func NewRecord(userID int64, familyMembers []int64) *Record {
	ts := time.Now()
	var one Record
	if err := gofakeit.Struct(&one); err != nil {
		log.Fatalln(err)
	}
	one.CreatedAt = ts.Unix()
	one.UpdatedAt = ts.Unix()
	one.UserID = userID
	// pin
	one.Pin = RandomPIN(0)
	// random others
	one.Others = RandomOthers()
	// random family shares
	if len(familyMembers) > 0 {
		if mathrand.Intn(2) == 1 {
			one.SharedAt = ts.Unix()
			one.SharedToAll = mathrand.Intn(2)
			var sharingMembers []int64
			for _, member := range familyMembers {
				if mathrand.Intn(2) == 1 {
					sharingMembers = append(sharingMembers, member)
				}
			}
			if one.SharedToAll != 1 && len(sharingMembers) > 0 {
				data, err := json.Marshal(sharingMembers)
				if err != nil {
					log.Fatalln(err)
				}
				one.SharingMembers = string(data)
			}
		}
	}
	switch one.RecordType {
	case "identity":
		one.FirstName = gofakeit.FirstName()
		one.LastName = gofakeit.LastName()
		one.FullName = one.FirstName + " " + one.LastName
		one.Job = gofakeit.JobTitle()
		one.SocialSecurityNumber = gofakeit.SSN()
		one.IdNumber = gofakeit.SSN()
	case "credit card":
		one.FullName = gofakeit.Name()
		one.CardholderName = one.FullName
		one.VerificationNumber = RandomPIN(4)
		one.ValidFrom = gofakeit.Date().Format("2006-01-02")
		one.IssuingBank = gofakeit.BookTitle()
	case "bank account":
		one.FullName = gofakeit.Name()
		one.BankName = gofakeit.BookTitle()
		one.NameOnAccount = one.FullName
		one.RoutingNumber = RandomPIN(8)
		one.Branch = gofakeit.BookTitle()
		one.AccountNumber = RandomPIN(32)
		one.Swift = RandomPIN(24)
	case "driver license":
		one.FullName = gofakeit.Name()
		one.Height = RandomPIN(3)
		one.LicenseClass = gofakeit.Car().Type
		one.State = gofakeit.State()
		one.Country = gofakeit.Country()
	case "passport":
		one.FullName = gofakeit.Name()
		one.IssuingCountry = gofakeit.Country()
		one.Nationality = gofakeit.Country()
		one.IssuingAuthority = gofakeit.BookTitle()
		one.BirthPlace = gofakeit.Address().Address
		one.IssuedOn = gofakeit.Date().Format("2006-01-02")
	}
	return &one
}

func InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx context.Context, userID int64, familyID string, familyMembers []int64, length int) {
	otherMembers := excludeMyself(userID, familyMembers)
	var nonPasswordRecords []*NonPasswordRecord
	var familySharedRecords []*FamilySharedRecord
	for i := 0; i < length; i++ {
		nonPassword := NewRecord(userID, otherMembers)
		data, err := json.Marshal(xorRecord(_initSlarkSecondaryPasswordHash, *nonPassword))
		if err != nil {
			log.Fatalln(err)
		}
		nonPasswordRecords = append(nonPasswordRecords, &NonPasswordRecord{
			DataID:  nonPassword.DataID,
			UserID:  userID,
			Type:    nonPassword.RecordType,
			Content: EncryptByUnlockPassword(data),
			Version: 1,
		})
		if nonPassword.SharedAt > 0 {
			content, err := EncryptByFamilyKey(_initFamilyKey, data)
			if err != nil {
				log.Fatalln(err)
			}
			familySharedRecords = append(familySharedRecords, &FamilySharedRecord{
				DataID:         nonPassword.DataID,
				FamilyID:       familyID,
				SharedBy:       nonPassword.UserID,
				Type:           nonPassword.RecordType,
				Content:        string(content),
				SharedToAll:    int32(nonPassword.SharedToAll),
				SharingMembers: nonPassword.SharingMembers,
				Version:        1,
			})
		}
	}
	if err := _pswdsDaoManager.NonPasswordRecordDAO.Create(ctx, nonPasswordRecords); err != nil {
		log.Fatalln(err)
	}
	if len(familySharedRecords) > 0 {
		if err := _pswdsDaoManager.FamilySharedRecordDAO.Create(ctx, familySharedRecords); err != nil {
			log.Fatalln(err)
		}
	}
}

// utils

func EncryptByFamilyKey(familyKey []byte, data []byte) ([]byte, error) {
	_, aead, err := simplecrypto.NewNonceAndX(familyKey[:32])
	if err != nil {
		return nil, err
	}
	return simplecrypto.EncryptByX(data, aead, simplecrypto.NonceZeroX()), nil
}

func DecryptFamilyKey(
	encryptedFamilyKey []byte,
	userPrivKey *ecies.PrivateKey,
) ([]byte, error) {
	return EciesDecrypt(
		userPrivKey,
		encryptedFamilyKey,
	)
}

func EncryptedFamilyKey(
	familyKey []byte,
	userPrivKey *ecies.PrivateKey,
) ([]byte, error) {
	return EciesEncrypt(userPrivKey.PublicKey, familyKey)
}

func FamilyKey(unlockPassword string) ([]byte, error) {
	return simplecrypto.Keccak256([]byte(fmt.Sprintf("%s%dC6093FD9CC143F9F058938868B2DF2DAF9A91D28", unlockPassword, time.Now().UnixMilli())))
}

func UserPrivateKey(unlockPassword string) (*ecies.PrivateKey, error) {
	hexKey, err := simplecrypto.Keccak256Hex([]byte(unlockPassword + "4838B106FCE9647BDF1E7877BF73CE8B0BAD5F97"))
	if err != nil {
		return nil, err
	}
	return ecies.NewPrivateKeyFromHex(string(hexKey))
}

func EciesEncrypt(pubKey *ecies.PublicKey, data []byte) ([]byte, error) {
	return ecies.Encrypt(pubKey, data)
}

func EciesDecrypt(privKey *ecies.PrivateKey, data []byte) ([]byte, error) {
	return ecies.Decrypt(privKey, data)
}

// slark account records

func insertSlarkData(ctx context.Context) {
	for _, email := range _initSlarkEmails {
		// 1. delete old account
		var oldAccount SlkUser
		if err := _slarkGormDB.Table(TableNameSlkUser).Select("id").First(&oldAccount, `email = ?`, email).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				log.Fatalln(err)
			}
		} else {
			if oldAccount.ID > 0 {
				needToDeletedUserIDs = append(needToDeletedUserIDs, oldAccount.ID)
			}
			if err := _slarkGormDB.Table(TableNameSlkUser).
				Delete(&oldAccount, `email = ?`, email).Error; err != nil {
				log.Fatalln(err)
			}
		}
		// 2. insert new account
		account := SlkUser{
			Email:                 email,
			Nickname:              email,
			PasswordHash:          _initSlarkPasswordHash,
			SecondaryPasswordHash: _initSlarkSecondaryPasswordHash,
		}
		if err := _slarkGormDB.Table(TableNameSlkUser).
			Create(&account).Error; err != nil {
			log.Fatalln(err)
		}
		_initSlarkAccounts[email] = &account
	}
	log.Println("=================> slark accounts all inserted successfully !!!")
}

func InsertTrustedContacts(ctx context.Context, userID int64) {
	var trustedContacts []TrustedContact
	encryptKey, err := simplecrypto.Keccak256([]byte(_initTrustedContactPassword + "9C9B913EB1B6254F4737CE947EFD16F16E916F"))
	if err != nil {
		log.Fatalln(err)
	}
	_, aead, err := simplecrypto.NewNonceAndX(encryptKey[:32])
	if err != nil {
		log.Fatalln(err)
	}
	for range 3 {
		trustedContacts = append(trustedContacts, TrustedContact{
			UserID:           userID,
			ContactEmail:     gofakeit.Email(),
			BackupCiphertext: string(simplecrypto.EncryptByX([]byte(_initSlarkSecondaryPassword), aead, simplecrypto.NonceZeroX())),
		})
	}
	if err := _pswdsDaoManager.TrustedContactDAO.Create(ctx, trustedContacts); err != nil {
		log.Fatalln(err)
	}
}

func insertPswdsData(ctx context.Context) {
	// 1. delete old data
	// 1-1. delete old backup records
	if err := _pswdsGormDB.
		Table(TableNameBackup).
		Delete(&Backup{}, `user_id IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-2. delete password records
	if err := _pswdsGormDB.
		Table(TableNamePasswordRecord).
		Delete(&PasswordRecord{}, `user_id IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-3. delete non password records
	if err := _pswdsGormDB.
		Table(TableNameNonPasswordRecord).
		Delete(&NonPasswordRecord{}, `user_id IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-4. delete family records
	if err := _pswdsGormDB.
		Table(TableNameFamily).
		Delete(&Family{}, `created_by IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-5. delete family member records
	if err := _pswdsGormDB.
		Table(TableNameFamilyMember).
		Delete(&FamilyMember{}, `user_id IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-6. delete family invitation records
	if err := _pswdsGormDB.
		Table(TableNameFamilyInvitation).
		Delete(&FamilyInvitation{}, `invited_by IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-7. delete family shared record records
	if err := _pswdsGormDB.
		Table(TableNameFamilySharedRecord).
		Delete(&FamilySharedRecord{}, `shared_by IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-8. delete family message records
	if err := _pswdsGormDB.
		Table(TableNameFamilyMessage).
		Delete(&FamilyMessage{}, `created_by IN (?)`, _initSlarkEmails).Error; err != nil {
		log.Fatalln(err)
	}
	// 1-9. delete trusted contact records
	if err := _pswdsGormDB.
		Table(TableNameTrustedContact).
		Delete(&TrustedContact{}, `user_id IN (?)`, needToDeletedUserIDs).Error; err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> delete pswds old data successfully !!!")
	// 2. insert new data
	var backups []Backup
	var families []Family
	var familyMembers []FamilyMember
	var familyInvitations []FamilyInvitation
	firstFamilyID := random.NewUUIDString()
	var firstFamilyCreator int64
	secondFamilyID := random.NewUUIDString()
	var secondFamilyCreator int64
	for _, account := range _initSlarkAccounts {
		backup := Backup{
			UserID:                      account.ID,
			PasswordHash:                account.SecondaryPasswordHash,
			UserPublicKey:               _initUserPrivKey.PublicKey.Hex(false),
			SecurityQuestions:           _initSecurityQuestions,
			SecurityQuestionsCiphertext: _initSecurityQuestionsCiphertext,
			// EncryptedFamilyKey:          _initEncryptedFamilyKey,
		}
		switch account.Email {
		case "test001@gmail.com": // first family -- creator
			families = append(families, Family{
				CreatedBy:   account.ID,
				FamilyID:    firstFamilyID,
				Description: "first family",
			})
			firstFamilyCreator = account.ID
			_firstFamilyMembers = append(_firstFamilyMembers, account.ID)
			familyMembers = append(familyMembers, FamilyMember{
				UserID:   account.ID,
				FamilyID: firstFamilyID,
				IsAdmin:  dao.FamilyMemberIsAdmin,
			})
			backup.EncryptedFamilyKey = _initEncryptedFamilyKey
		case "test002@gmail.com": // first family -- not admin
			_firstFamilyMembers = append(_firstFamilyMembers, account.ID)
			familyMembers = append(familyMembers, FamilyMember{
				UserID:   account.ID,
				FamilyID: firstFamilyID,
				IsAdmin:  dao.FamilyMemberIsNotAdmin,
			})
			backup.EncryptedFamilyKey = _initEncryptedFamilyKey
		case "test003@gmail.com": // no family -- has first family invitation
			familyInvitations = append(familyInvitations, FamilyInvitation{
				FamilyID:           firstFamilyID,
				InvitedBy:          firstFamilyCreator,
				Email:              account.Email,
				EncryptedFamilyKey: _initEncryptedFamilyKey,
			})
		case "test004@gmail.com": // no family -- no family invitation
		case "test005@gmail.com": // second family -- creator
			families = append(families, Family{
				CreatedBy:   account.ID,
				FamilyID:    secondFamilyID,
				Description: "second family",
			})
			secondFamilyCreator = account.ID
			_secondFamilyMembers = append(_secondFamilyMembers, account.ID)
			familyMembers = append(familyMembers, FamilyMember{
				UserID:   account.ID,
				FamilyID: secondFamilyID,
				IsAdmin:  dao.FamilyMemberIsAdmin,
			})
			backup.EncryptedFamilyKey = _initEncryptedFamilyKey
		case "test006@gmail.com": // second family -- admin
			_secondFamilyMembers = append(_secondFamilyMembers, account.ID)
			familyMembers = append(familyMembers, FamilyMember{
				UserID:   account.ID,
				FamilyID: secondFamilyID,
				IsAdmin:  dao.FamilyMemberIsAdmin,
			})
			backup.EncryptedFamilyKey = _initEncryptedFamilyKey
		case "test007@gmail.com": // second family -- not admin
			_secondFamilyMembers = append(_secondFamilyMembers, account.ID)
			familyMembers = append(familyMembers, FamilyMember{
				UserID:   account.ID,
				FamilyID: secondFamilyID,
				IsAdmin:  dao.FamilyMemberIsNotAdmin,
			})
			backup.EncryptedFamilyKey = _initEncryptedFamilyKey
		case "test008@gmail.com": // no family -- has second family invitation
			familyInvitations = append(familyInvitations, FamilyInvitation{
				FamilyID:           secondFamilyID,
				InvitedBy:          secondFamilyCreator,
				Email:              account.Email,
				EncryptedFamilyKey: _initEncryptedFamilyKey,
			})
		case "test009@gmail.com": // no family -- no family invitation
		}
		backups = append(backups, backup)
	}

	// 2-1. insert backups
	if err := _pswdsDaoManager.BackupDAO.Create(ctx, backups); err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> insert pswds backup records successfully !!!")
	// 2-2. insert families
	if err := _pswdsDaoManager.FamilyDAO.Create(ctx, families); err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> insert pswds family records successfully !!!")
	// 2-3. insert family members
	if err := _pswdsDaoManager.FamilyMemberDAO.Create(ctx, familyMembers); err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> insert pswds family member records successfully !!!")
	// 2-4. insert family invitations
	if err := _pswdsDaoManager.FamilyInvitationDAO.Create(ctx, familyInvitations); err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> insert pswds family invitations records successfully !!!")
	// 2-5. insert family messages
	var familyMessages []FamilyMessage
	for range 100 {
		var familyID string
		var createdBy string
		if mathrand.Intn(2) == 1 {
			familyID = secondFamilyID
			createdBy = _initSlarkEmails[0]
		} else {
			familyID = firstFamilyID
			createdBy = _initSlarkEmails[4]
		}
		var target string
		targetIndex := mathrand.Intn(8)
		if familyID == firstFamilyID {
			target = _initSlarkEmails[1:][targetIndex]
		} else {
			otherMembers := _initSlarkEmails[:4]
			otherMembers = append(otherMembers, _initSlarkEmails[5:]...)
			target = otherMembers[targetIndex]
		}
		familyMessages = append(familyMessages, FamilyMessage{
			FamilyID:  familyID,
			CreatedBy: createdBy,
			Target:    target,
			Operation: int32(mathrand.Intn(8) + 1),
		})
	}
	if err := _pswdsDaoManager.FamilyMessageDAO.Create(ctx, familyMessages); err != nil {
		log.Fatalln(err)
	}
	log.Println("=================> insert pswds family message records successfully !!!")
	// 2-6. insert password, non password records, trusted contact records
	for _, account := range _initSlarkAccounts {
		passwordsLength := 100
		nonPasswordsLength := 100
		switch account.Email {
		case "test001@gmail.com": // first family -- creator
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, firstFamilyID, _firstFamilyMembers, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, firstFamilyID, _firstFamilyMembers, nonPasswordsLength)
		case "test002@gmail.com": // first family -- not admin
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, firstFamilyID, _firstFamilyMembers, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, firstFamilyID, _firstFamilyMembers, nonPasswordsLength)
		case "test003@gmail.com": // no family -- has first family invitation
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, "", nil, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, "", nil, nonPasswordsLength)
		case "test004@gmail.com": // no family -- no family invitation
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, "", nil, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, "", nil, nonPasswordsLength)
		case "test005@gmail.com": // second family -- creator
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, nonPasswordsLength)
		case "test006@gmail.com": // second family -- admin
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, nonPasswordsLength)
		case "test007@gmail.com": // second family -- not admin
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, secondFamilyID, _secondFamilyMembers, nonPasswordsLength)
		case "test008@gmail.com": // no family -- has second family invitation
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, "", nil, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, "", nil, nonPasswordsLength)
		case "test009@gmail.com": // no family -- no family invitation
			InsertPasswordsAndSharedPasswordsByUserID(ctx, account.ID, "", nil, passwordsLength)
			InsertNonPasswordsAndSharedNonPasswordsByUserID(ctx, account.ID, "", nil, nonPasswordsLength)
		}
		// trusted contact records
		InsertTrustedContacts(ctx, account.ID)
		time.Sleep(200 * time.Millisecond)
	}
	log.Println("=================> insert pswds all records successfully !!!")
}

// 每轮100个账号，每个账号 100 条password，100条 non password
func insertSlarkAccountsN(ctx context.Context, loop int64) {
	for range loop {
		var users []SlkUser
		for range 50 {
			one := SlkUser{
				Email:                 random.NewUUIDHexEncoding()[:8] + gofakeit.Email(),
				Nickname:              gofakeit.FirstName() + "-" + random.NewUUIDString()[:8],
				PasswordHash:          _initSlarkPasswordHash,
				SecondaryPasswordHash: _initSlarkSecondaryPasswordHash,
			}
			users = append(users, one)
		}
		// users
		if err := _slarkGormDB.Table(TableNameSlkUser).
			Create(users).Error; err != nil {
			log.Fatalln(err)
		}
		time.Sleep(10 * time.Millisecond)
		// pswds data
		var backups []Backup
		for _, user := range users {
			one := Backup{
				UserID:                      user.ID,
				PasswordHash:                _initSlarkSecondaryPasswordHash,
				UserPublicKey:               _initUserPrivKey.PublicKey.Hex(false),
				EncryptedFamilyKey:          _initEncryptedFamilyKey,
				SecurityQuestions:           _initSecurityQuestions,
				SecurityQuestionsCiphertext: _initSecurityQuestionsCiphertext,
			}
			backups = append(backups, one)
			var passwords []PasswordRecord
			var nonPasswords []NonPasswordRecord
			for range 50 {
				password := NewPassword(user.ID, nil)
				data, err := json.Marshal(xorPassword(_initSlarkSecondaryPasswordHash, *password))
				if err != nil {
					log.Fatalln(err)
				}
				one := PasswordRecord{
					DataID:  password.DataID,
					UserID:  user.ID,
					Content: EncryptByUnlockPassword(data),
					Version: 1,
				}
				passwords = append(passwords, one)
			}
			// passwords
			if err := _pswdsDaoManager.PasswordRecordDAO.Create(ctx, passwords); err != nil {
				log.Fatalln(err)
			}
			time.Sleep(10 * time.Millisecond)
			for range 50 {
				nonPassword := NewRecord(user.ID, nil)
				data, err := json.Marshal(xorRecord(_initSlarkSecondaryPasswordHash, *nonPassword))
				if err != nil {
					log.Fatalln(err)
				}
				one := NonPasswordRecord{
					DataID:  nonPassword.DataID,
					UserID:  user.ID,
					Type:    nonPassword.RecordType,
					Content: EncryptByUnlockPassword(data),
					Version: 1,
				}
				nonPasswords = append(nonPasswords, one)
			}
			// non passwords
			if err := _pswdsDaoManager.NonPasswordRecordDAO.Create(ctx, nonPasswords); err != nil {
				log.Fatalln(err)
			}
			time.Sleep(10 * time.Millisecond)
		}
		// backups
		if err := _pswdsDaoManager.BackupDAO.Create(ctx, backups); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	log.Printf("=================> %d accounts' data all inserted successfully !!!\n", loop*50)
}

func main() {
	ctx := context.Background()
	// 9个账号数据 -- 有家庭/家庭邀请的数据
	insertSlarkData(ctx)
	insertPswdsData(ctx)
	// 成千上万个账号数据 -- 只有password、non password数据
	// 500个账号数据量大概100M，该函数需要单独执行 10 次左右会有1G数据量
	insertSlarkAccountsN(ctx, 100)
}
