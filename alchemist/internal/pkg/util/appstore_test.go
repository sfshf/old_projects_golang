package util_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/internal/pkg/util"
)

func TestServerNotificationV2(t *testing.T) {
	// {"signedPayload":"..."}
	appStoreServerRequest := os.Getenv("APPLE_NOTIFICATION_REQUEST")
	if appStoreServerRequest == "" {
		t.Error("No valid AppStoreServerRequest")
	}
	var request util.AppStoreServerRequest
	err := json.Unmarshal([]byte(appStoreServerRequest), &request) // bind byte to header structure
	if err != nil {
		t.Error(err)
	}

	// -----BEGIN CERTIFICATE----- ......
	rootCert := os.Getenv("APPLE_CERT")
	if rootCert == "" {
		t.Error("Apple Root Cert not available")
	}

	appStoreServerNotification, err := util.New(request.SignedPayload, rootCert)
	if err != nil {
		t.Error(err)
	}

	if !appStoreServerNotification.IsValid {
		t.Error("Payload is not valid")
	}

	if appStoreServerNotification.Payload.Data.Environment != "sandbox" {
		t.Errorf("got %s, want sandbox", appStoreServerNotification.Payload.Data.Environment)
	}

	println(appStoreServerNotification.Payload.Data.BundleId)
	println(appStoreServerNotification.TransactionInfo.ProductId)
	fmt.Printf("Product Id: %s", appStoreServerNotification.RenewalInfo.ProductId)
}

func TestGeneratePromotionalOfferSignature(t *testing.T) {
	res, err := util.GeneratePromotionalOfferSignature(
		"",
		util.AppConfig("alchemist.test").AppID,
		util.AppConfig("alchemist.test").PromoOfferKeyID,
		util.AppConfig("alchemist.test").AppID,
		"",
		"",
		util.NewUUIDString(),
		fmt.Sprintf("%d", time.Now().UnixMilli()),
	)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(res)
}
