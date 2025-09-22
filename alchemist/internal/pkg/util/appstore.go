package util

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"strings"

	"github.com/golang-jwt/jwt"
)

type AppStoreServerNotification struct {
	appleRootCert   string
	Payload         *NotificationPayload
	TransactionInfo *TransactionInfo
	RenewalInfo     *RenewalInfo
	IsValid         bool
}

type AppStoreServerRequest struct {
	SignedPayload string `json:"signedPayload"`
}

type NotificationHeader struct {
	Alg string   `json:"alg"`
	X5c []string `json:"x5c"`
}

type NotificationPayload struct {
	jwt.StandardClaims
	NotificationType string              `json:"notificationType"`
	Subtype          string              `json:"subtype"`
	NotificationUUID string              `json:"notificationUUID"`
	Version          string              `json:"version"`
	Summary          NotificationSummary `json:"summary"`
	Data             NotificationData    `json:"data"`
}

type NotificationSummary struct {
	RequestIdentifier      string   `json:"requestIdentifier"`
	AppAppleId             string   `json:"appAppleId"`
	BundleId               string   `json:"bundleId"`
	ProductId              string   `json:"productId"`
	Environment            string   `json:"environment"`
	StoreFrontCountryCodes []string `json:"storefrontCountryCodes"`
	FailedCount            int64    `json:"failedCount"`
	SucceededCount         int64    `json:"succeededCount"`
}

type NotificationData struct {
	AppAppleId            int    `json:"appAppleId"`
	BundleId              string `json:"bundleId"`
	BundleVersion         string `json:"bundleVersion"`
	Environment           string `json:"environment"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
	Status                int32  `json:"status"`
}

type TransactionInfo struct {
	jwt.StandardClaims
	AppAccountToken             string `json:"appAccountToken"`
	BundleId                    string `json:"bundleId"`
	Currency                    string `json:"currency"`
	Environment                 string `json:"environment"`
	ExpiresDate                 int64  `json:"expiresDate"`
	InAppOwnershipType          string `json:"inAppOwnershipType"`
	IsUpgraded                  bool   `json:"isUpgraded"`
	OfferDiscountType           string `json:"offerDiscountType"`
	OfferIdentifier             string `json:"offerIdentifier"`
	OfferType                   int32  `json:"offerType"`
	OriginalPurchaseDate        int64  `json:"originalPurchaseDate"`
	OriginalTransactionId       string `json:"originalTransactionId"`
	Price                       int32  `json:"price"`
	ProductId                   string `json:"productId"`
	PurchaseDate                int64  `json:"purchaseDate"`
	Quantity                    int32  `json:"quantity"`
	RevocationDate              int64  `json:"revocationDate"`
	RevocationReason            int32  `json:"revocationReason"`
	SignedDate                  int64  `json:"signedDate"`
	Storefront                  string `json:"storefront"`
	StorefrontId                string `json:"storefrontId"`
	SubscriptionGroupIdentifier string `json:"subscriptionGroupIdentifier"`
	TransactionId               string `json:"transactionId"`
	TransactionReason           string `json:"transactionReason"`
	Type                        string `json:"type"`
	WebOrderLineItemId          string `json:"webOrderLineItemId"`
}

type RenewalInfo struct {
	jwt.StandardClaims
	AutoRenewProductId          string `json:"autoRenewProductId"`
	AutoRenewStatus             int32  `json:"autoRenewStatus"`
	Environment                 string `json:"environment"`
	ExpirationIntent            int32  `json:"expirationIntent"`
	GracePeriodExpiresDate      int64  `json:"gracePeriodExpiresDate"`
	IsInBillingRetryPeriod      bool   `json:"isInBillingRetryPeriod"`
	OfferIdentifier             string `json:"offerIdentifier"`
	OfferType                   int32  `json:"offerType"`
	OriginalTransactionId       string `json:"originalTransactionId"`
	PriceIncreaseStatus         int32  `json:"priceIncreaseStatus"`
	ProductId                   string `json:"productId"`
	RecentSubscriptionStartDate int64  `json:"recentSubscriptionStartDate"`
	RenewalDate                 int64  `json:"renewalDate"`
	SignedDate                  int64  `json:"signedDate"`
}

func New(payload string, appleRootCert string) (*AppStoreServerNotification, error) {
	asn := &AppStoreServerNotification{}
	asn.IsValid = false
	asn.appleRootCert = appleRootCert
	if err := asn.parseJwtSignedPayload(payload); err != nil {
		return nil, err
	}
	return asn, nil
}

func (asn *AppStoreServerNotification) extractHeaderByIndex(payload string, index int) ([]byte, error) {
	// get header from token
	payloadArr := strings.Split(payload, ".")

	// convert header to byte
	headerByte, err := base64.RawStdEncoding.DecodeString(payloadArr[0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// bind byte to header structure
	var header NotificationHeader
	err = json.Unmarshal(headerByte, &header)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// decode x.509 certificate headers to byte
	certByte, err := base64.StdEncoding.DecodeString(header.X5c[index])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return certByte, nil
}

func (asn *AppStoreServerNotification) verifyCertificate(certByte []byte, intermediateCert []byte) error {
	// create certificate pool
	roots := x509.NewCertPool()

	// parse and append apple root certificate to the pool
	ok := roots.AppendCertsFromPEM([]byte(asn.appleRootCert))
	if !ok {
		log.Println("root certificate couldn't be parsed")
		return errors.New("root certificate couldn't be parsed")
	}

	// parse and append intermediate x5c certificate
	interCert, err := x509.ParseCertificate(intermediateCert)
	if err != nil {
		log.Println("intermediate certificate couldn't be parsed")
		return errors.New("intermediate certificate couldn't be parsed")
	}
	intermediate := x509.NewCertPool()
	intermediate.AddCert(interCert)

	// parse x5c certificate
	cert, err := x509.ParseCertificate(certByte)
	if err != nil {
		log.Println(err)
		return err
	}

	// verify X5c certificate using app store certificate resides in opts
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediate,
	}
	if _, err := cert.Verify(opts); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (asn *AppStoreServerNotification) extractPublicKeyFromPayload(payload string) (*ecdsa.PublicKey, error) {
	// get certificate from X5c[0] header
	certStr, err := asn.extractHeaderByIndex(payload, 0)
	if err != nil {
		return nil, err
	}

	// parse certificate
	cert, err := x509.ParseCertificate(certStr)
	if err != nil {
		return nil, err
	}

	// get public key
	switch pk := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		return pk, nil
	default:
		return nil, errors.New("appstore public key must be of type ecdsa.PublicKey")
	}
}

func (asn *AppStoreServerNotification) parseJwtSignedPayload(payload string) error {
	// get root certificate from x5c header
	rootCertStr, err := asn.extractHeaderByIndex(payload, 2)
	// log.Println("rootCertStr:", rootCertStr)
	if err != nil {
		return err
	}

	// get intermediate certificate from x5c header
	intermediateCertStr, err := asn.extractHeaderByIndex(payload, 1)
	// log.Println("intermediateCertStr:", intermediateCertStr)
	if err != nil {
		return err
	}

	// verify certificates
	if err = asn.verifyCertificate(rootCertStr, intermediateCertStr); err != nil {
		return err
	}

	// payload data
	notificationPayload := &NotificationPayload{}
	_, err = jwt.ParseWithClaims(payload, notificationPayload, func(token *jwt.Token) (interface{}, error) {
		return asn.extractPublicKeyFromPayload(payload)
	})
	if err != nil {
		return err
	}
	// log.Printf("NotificationPayload: %#v\n", *notificationPayload)
	asn.Payload = notificationPayload

	// transaction info
	payload = asn.Payload.Data.SignedTransactionInfo
	transactionInfo := &TransactionInfo{}
	if payload != "" {
		// log.Println("signedTransactionInfo:", payload)
		_, err = jwt.ParseWithClaims(payload, transactionInfo, func(token *jwt.Token) (interface{}, error) {
			return asn.extractPublicKeyFromPayload(payload)
		})
		if err != nil {
			return err
		}
		// log.Printf("TransactionInfo: %#v\n", *transactionInfo)
	}
	asn.TransactionInfo = transactionInfo

	// renewal info
	payload = asn.Payload.Data.SignedRenewalInfo
	renewalInfo := &RenewalInfo{}
	if payload != "" {
		// log.Println("signedRenewalInfo:", payload)
		_, err = jwt.ParseWithClaims(payload, renewalInfo, func(token *jwt.Token) (interface{}, error) {
			return asn.extractPublicKeyFromPayload(payload)
		})
		if err != nil {
			return err
		}
		// log.Printf("RenewalInfo: %#v\n", *renewalInfo)
	}
	asn.RenewalInfo = renewalInfo

	// valid request
	asn.IsValid = true
	return nil
}

func CombineOfferMessage(appBundleId, keyID, productIdentifier, offerIdentifier, appAccountToken, nonce, timestamp string) (string, error) {
	var b strings.Builder
	if _, err := b.WriteString(appBundleId); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	if _, err := b.WriteString(keyID); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	if _, err := b.WriteString(productIdentifier); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	if _, err := b.WriteString(offerIdentifier); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	if _, err := b.WriteString(appAccountToken); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	// nonce
	if _, err := b.WriteString(nonce); err != nil {
		return "", err
	}
	if _, err := b.WriteRune('\u2063'); err != nil {
		return "", err
	}
	if _, err := b.WriteString(timestamp); err != nil {
		return "", err
	}
	return b.String(), nil
}

func buildPrivateKey(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}
	anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := anyKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to parse PKCS8 private key")
	}
	return privateKey, nil
}

func GeneratePromotionalOfferSignature(appName, appBundleId, keyID, productIdentifier, offerIdentifier, appAccountToken, nonce, timestamp string) (string, error) {
	combined, err := CombineOfferMessage(appBundleId, keyID, productIdentifier, offerIdentifier, appAccountToken, nonce, timestamp)
	if err != nil {
		return "", err
	}
	privKey, err := buildPrivateKey([]byte(AppConfig(appName).PromoOfferPrivKeyPem))
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(combined))
	sig, err := ecdsa.SignASN1(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}
