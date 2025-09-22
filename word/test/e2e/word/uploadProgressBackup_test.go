package word_test

import (
	"bytes"
	"context"
	"io"
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

func TestUploadProgressBackup(t *testing.T) {
	ctx := context.Background()
	// mock data
	testData := "word e2e testing TestUploadProgressBackup."

	testTS := time.Now()
	mockBackup := ProgressBackup{
		UserID:    _testAccount.ID,
		Version:   1,
		Timestamp: testTS.UnixMilli(),
		Resource:  "/path/to/resource",
	}
	if err := _wordGormDB.Create(&mockBackup).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _wordGormDB.Delete(&ProgressBackup{ID: mockBackup.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		Timestamp int64  `json:"timestamp"`
		Version   int32  `json:"version"`
		Data      string `json:"data"`
	}{
		Timestamp: testTS.Add(5 * time.Minute).UnixMilli(),
		Version:   1,
		Data:      testData,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/user/progress/backup/upload/v1", &reqData, _testCookie, &respData, nil)
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
	var backupLog ProgressBackup
	if err := _wordGormDB.Table(TableNameProgressBackup).
		Where(`user_id = ? AND timestamp = ? AND version = ? AND deleted_at = 0`,
			_testAccount.ID, testTS.Add(5*time.Minute).UnixMilli(), 1).
		First(&backupLog).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _wordGormDB.Delete(&ProgressBackup{ID: backupLog.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	getObjectOutput, err := _s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(_progressBackupBucketName),
		Key:    aws.String(backupLog.Resource),
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := _s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(_progressBackupBucketName),
			Key:    aws.String(backupLog.Resource),
		}); err != nil {
			t.Error(err)
			return
		}
	}()
	body, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		t.Error(err)
		return
	}
	plaintext, err := util.AES16CBCDecrypt(body, []byte(os.Getenv("STUDY_BACKUP_KEY")))
	if err != nil {
		t.Error(err)
		return
	}
	if sum1, sum2 := md5Sum(plaintext), md5Sum([]byte(testData)); !bytes.Equal(sum1, sum2) {
		t.Error("not prospective response data")
		return
	}
}
