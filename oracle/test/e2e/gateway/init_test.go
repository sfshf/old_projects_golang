package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_oracleMysqlDsn = os.Getenv("GATEWAY_MYSQL_DNS") + "?charset=utf8&parseTime=true"
	_oracleGormDB   *gorm.DB

	_oracleConsoleDNS    = os.Getenv("ORACLE_CONSOLE_DNS")
	_oracleGwDNS         = os.Getenv("ORACLE_DNS")
	_oracleConsoleApiKey = os.Getenv("ORACLE_CONSOLE_APIKEY")
	_oracleGatewayApiKey = os.Getenv("ORACLE_GATEWAY_APIKEY")

	_uploadBucketName = "n1xt-upload"
	_s3Client         *s3.Client
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("CONSUL_HTTP_ADDR", "172.31.29.192:8500")
	var err error
	// mysql
	_oracleGormDB, err = gorm.Open(mysql.Open(_oracleMysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// aws
	os.Setenv("AWS_REGION", "us-east-1")
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

func postJsonRequest(location string, reqData interface{}, cookie *http.Cookie, respData interface{}, reqHeaderFunc func(req *http.Request)) (*http.Response, error) {
	log.Printf("location: %s\n", location)
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return nil, err
		}
		log.Printf("request data: %s\n", jsonData)
		body = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(http.MethodPost, location, body)
	if err != nil {
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
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("response data: %s\n", data)
	if respData != nil {
		if err := json.Unmarshal(data, respData); err != nil {
			return resp, err
		}
	}
	return resp, nil
}
