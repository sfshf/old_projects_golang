package word_test

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nextsurfer/word/api/response"
	. "github.com/nextsurfer/word/internal/pkg/model"
	"github.com/nextsurfer/word/internal/pkg/util"
)

func TestDownloadProgressBackup(t *testing.T) {
	ctx := context.Background()
	// mock data
	testData := "this is a test data from word e2e testing TestDownloadProgressBackup."
	testEncryptedData, err := util.AES16CBCEncrypt([]byte(testData), []byte(os.Getenv("STUDY_BACKUP_KEY")))
	if err != nil {
		t.Error(err)
		return
	}
	testTS := time.Now()
	testS3Key := getHashedPath(1, _testAccount.ID, testTS.UnixMilli(), testEncryptedData)
	if _, err = _s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(_progressBackupBucketName),
		Key:    aws.String(testS3Key),
		Body:   bytes.NewReader(testEncryptedData),
	}); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := _s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(_progressBackupBucketName),
			Key:    aws.String(testS3Key),
		}); err != nil {
			t.Error(err)
			return
		}
	}()
	mockBackup := ProgressBackup{
		UserID:    _testAccount.ID,
		Version:   1,
		Timestamp: testTS.UnixMilli(),
		Resource:  testS3Key,
	}
	if err := _wordGormDB.Create(&mockBackup).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if mockBackup.ID > 0 {
			if err := _wordGormDB.Delete(&ProgressBackup{ID: mockBackup.ID}).Error; err != nil {
				t.Error(err)
				return
			}
		}
	}()

	reqData := struct {
		Timestamp int64 `json:"timestamp"`
		Version   int32 `json:"version"`
	}{
		Timestamp: testTS.Add(-5 * time.Minute).UnixMilli(),
		Version:   1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Content string `json:"content"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/user/progress/backup/download/v1", &reqData, _testCookie, &respData, nil)
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
	// check data
	if sum1, sum2 := md5Sum([]byte(respData.Data.Content)), md5Sum([]byte(testData)); !bytes.Equal(sum1, sum2) {
		t.Error("not prospective response data")
		return
	}
}
