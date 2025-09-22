package doom_console_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_kongDNS           = "https://api.test.n1xt.net"
	_doomConsoleApiKey = "m1IaZ/gSdTUPgVECk/+BiBC8"

	_mongoDB *mongo.Database
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("CONSUL_HTTP_ADDR", "172.31.29.192:8500")

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
