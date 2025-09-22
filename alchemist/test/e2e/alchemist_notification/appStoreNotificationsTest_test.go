package alchemistnotification_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
)

// func TestExportRawTransactions(t *testing.T) {
// 	var rawTransactions []*RawTransaction
// 	if err := _alchemistGormDB.Table("raw_transactions").Find(&rawTransactions).Error; err != nil {
// 		log.Println(err)
// 	}
// 	f, err := os.Create("raw_transactions")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	defer f.Close()
// 	var off int64
// 	for _, rawTransaction := range rawTransactions {
// 		n, err := f.WriteAt([]byte(rawTransaction.Data+"\n"), off)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		off += int64(n)
// 	}
// 	if err := f.Sync(); err != nil {
// 		log.Println(err)
// 	}
// 	log.Println("ok")
// }

// func TestUnmarshalRawTransaction(t *testing.T) {
// 	f, err := os.Open("raw_transactions")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	defer f.Close()
// 	raw, err := io.ReadAll(f)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	raws := bytes.Split(raw, []byte("\n"))
// 	var reqDatas []struct {
// 		SignedPayload string `json:"signedPayload"`
// 	}
// 	for _, raw := range raws {
// 		if strings.TrimSpace(string(raw)) == "" {
// 			continue
// 		}
// 		var reqData struct {
// 			SignedPayload string `json:"signedPayload"`
// 		}
// 		fmt.Println(string(raw))
// 		if err := json.Unmarshal(bytes.TrimSpace(raw), &reqData); err != nil {
// 			t.Error(err)
// 			return
// 		}
// 		reqDatas = append(reqDatas, reqData)
// 	}
// }

func TestAppStoreNotificationTest(t *testing.T) {
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
			t.Error(err)
			return
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
		resp, err := postJsonRequest(_kongDNS+"/alchemist/appstore/test-notifications", &reqData, nil, &respData, nil)
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

	var transactionsTestCount int64
	if err := _alchemistGormDB.Table(TableNameTransactionsTest).
		Where("user_id = ? AND deleted_at = 0", accountToken.UserID).
		Count(&transactionsTestCount).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&TransactionsTest{}, "1=1").Error; err != nil {
			t.Error(err)
			return
		}
	}()
	if transactionsTestCount != 13 {
		t.Errorf("transactionsTestCount [%d] not equal to 13", transactionsTestCount)
		return
	}
	var subscriptionStateTest SubscriptionStateTest
	if err := _alchemistGormDB.Table(TableNameSubscriptionStateTest).
		Where("user_id = ? AND deleted_at = 0", accountToken.UserID).
		First(&subscriptionStateTest).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&subscriptionStateTest).Error; err != nil {
			t.Error(err)
			return
		}
	}()
}
