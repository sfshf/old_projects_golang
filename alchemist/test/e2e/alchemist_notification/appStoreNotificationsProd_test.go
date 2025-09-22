package alchemistnotification_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
)

func TestAppStoreNotificationProd(t *testing.T) {
	accountToken := SlarkUser{
		AppAccountToken: "5ca011a3-f1b6-417d-a0ba-061797ecb39c",
		UserID:          100000163,
	}
	if err := _alchemistGormDB.Create(&accountToken).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&accountToken).Error; err != nil {
			log.Println(err)
		}
	}()

	// test notification apis
	f, err := os.Open("raw_transactions")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
		return
	}
	raws := bytes.Split(raw, []byte("\n"))
	var reqDatas []struct {
		SignedPayload string `json:"signedPayload"`
	}
	for _, raw := range raws {
		if strings.TrimSpace(string(raw)) == "" {
			continue
		}
		var reqData struct {
			SignedPayload string `json:"signedPayload"`
		}
		if err := json.Unmarshal(raw, &reqData); err != nil {
			t.Error(err)
			return
		}
		reqDatas = append(reqDatas, reqData)
	}
	for _, reqData := range reqDatas {
		time.Sleep(1 * time.Second)
		respData := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		resp, err := postJsonRequest(_kongDNS+"/alchemist/appstore/notifications", &reqData, nil, &respData, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respData.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
	}

	var rawTransactionsCount int64
	if err := _alchemistGormDB.Table(TableNameRawTransaction).
		Count(&rawTransactionsCount).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&RawTransaction{}, "1=1").Error; err != nil {
			t.Error(err)
			return
		}
	}()
	if rawTransactionsCount != 13 {
		t.Errorf("rawTransactionsCount [%d] not equal to 13", rawTransactionsCount)
		return
	}

	// test handleAppStoreNotification
	time.Sleep(time.Second * 90)

	var transactionsProdCount int64
	if err := _alchemistGormDB.Table(TableNameTransactionsProd).
		Where("user_id = ? AND deleted_at = 0", accountToken.UserID).
		Count(&transactionsProdCount).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&TransactionsProd{}, "1=1").Error; err != nil {
			t.Error(err)
			return
		}
	}()
	if transactionsProdCount != 13 {
		t.Errorf("transactionsProdCount [%d] not equal to 13", transactionsProdCount)
		return
	}
	var subscriptionStateProd SubscriptionStateProd
	if err := _alchemistGormDB.Table(TableNameSubscriptionStateProd).
		Where("user_id = ? AND deleted_at = 0", accountToken.UserID).
		First(&subscriptionStateProd).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&subscriptionStateProd).Error; err != nil {
			t.Error(err)
			return
		}
	}()
}
