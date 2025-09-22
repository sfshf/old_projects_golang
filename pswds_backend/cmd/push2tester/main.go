package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/pswds_backend/internal/common/simplehttp"
)

var oracleGatewayDNS string

func init() {
	flag.StringVar(&oracleGatewayDNS, "dns", "https://api.test.n1xt.net", "oracle gateway dns")
}

type ApiTestcase struct {
	Name string
	Path string
	Body string
}

var apiTestcases = []ApiTestcase{
	{
		Name: "CheckPasswordBackup",
		Path: "/pswds/checkPasswordBackup/v1",
		Body: `{ "updatedAt": 78945612388 }`,
	},
	{
		Name: "UploadPasswordBackup",
		Path: "/pswds/uploadPasswordBackup/v1",
		Body: `{ "lockScreenPassword": "3333585", list:'[]' }`,
	},
	{
		Name: "DownloadPasswordBackup",
		Path: "/pswds/downloadPasswordBackup/v1",
		Body: ``,
	},
	{
		Name: "CreatePassword",
		Path: "/pswds/createPassword/v1",
		Body: `{ "id": "uuid", 
		"createdAt":78945612388, 
		"updatedAt":78945612388, 
		"title":"title", 
		"url":"url", 
		"usernameOrEmail":"usernameOrEmail",
		"password":"password",
		"version":1 }`,
	},
	{
		Name: "UpdatePassword",
		Path: "/pswds/updatePassword/v1",
		Body: `{ "id": "uuid", 
		"updatedAt":78945612374, 
		"title":"title", 
		"url":"url", 
		"usernameOrEmail":"usernameOrEmail",
		"password":"password",
		"version":1 }`,
	},
	{
		Name: "DeletePassword",
		Path: "/pswds/deletePassword/v1",
		Body: `{ "id": "", "version": 1 }`,
	},
}

func UpdateAPITestcase() error {
	data, err := json.Marshal(apiTestcases)
	if err != nil {
		return err
	}
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Data   string `json:"data"`
	}{
		ApiKey: "ZWrSgH9D6lrtoecZ7HGslpLo",
		App:    "pswds_backend",
		Data:   string(data),
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err := simplehttp.PostJsonRequest(oracleGatewayDNS+"/tester/updateAPITestcase/v1", &reqData, nil, &respData)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("UpdateAPITestcase: http code not equal to 200")
	}
	if respData.Code != 0 {
		return fmt.Errorf("UpdateAPITestcase: %v", respData)
	}
	return nil
}

func main() {
	if err := UpdateAPITestcase(); err != nil {
		log.Fatalln(err)
	}
}
