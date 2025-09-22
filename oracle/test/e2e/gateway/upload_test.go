package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestUploadFile(t *testing.T) {
	fileContent := []byte("This is the file content.")
	req, err := http.NewRequest(http.MethodPost, _oracleGwDNS+"/upload", bytes.NewReader(fileContent))
	if err != nil {
		t.Error(err)
		return
	}
	// req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			HashName string `json:"hashName"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(data, &respData); err != nil {
		t.Error(err)
		return
	}
	log.Printf("response data: %s\n", data)
	if respData.Code != 0 {
		t.Error("not prospective response data code")
		return
	}
	if respData.Data.HashName == "" {
		t.Error("not prospective response data")
		return
	}
	ctx := context.Background()
	getObjectOutput, err := _s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(_uploadBucketName),
		Key:    aws.String("upload/" + respData.Data.HashName),
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := _s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(_uploadBucketName),
			Key:    aws.String("upload/" + respData.Data.HashName),
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
	if string(body) != string(fileContent) {
		t.Error("not prospective result")
		return
	}
}
