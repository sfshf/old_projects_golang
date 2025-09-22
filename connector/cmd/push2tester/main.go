package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/connector/internal/pkg/simplehttp"
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
		Name: "CreatePassword",
		Path: "/riki/createPassword/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"pswd_TestConnectorPassword_lurgvep6t7v9dnlx"}`,
	},
	{
		Name: "CheckPassword",
		Path: "/riki/checkPassword/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"pswd_TestConnectorPassword_lurgvep6t7v9dnlx","passwordHash":"479beb3587b35e4095549f06d950f43da0c9c659e2d6fcf5aa116bc0fa4e1429"}`,
	},
	{
		Name: "DeletePassword",
		Path: "/riki/deletePassword/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"pswd_TestConnectorPassword_lurgvep6t7v9dnlx"}`,
	},
	{
		Name: "CreatePrivateKey",
		Path: "/riki/createPrivateKey/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorPrivateKey_c3k5r8"}`,
	},
	{
		Name: "CheckKeyExisting",
		Path: "/riki/checkKeyExisting/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorPrivateKey_c3k5r8"}`,
	},
	{
		Name: "GetPublicKey",
		Path: "/riki/getPublicKey/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorPrivateKey_c3k5r8"}`,
	},
	{
		Name: "DeletePrivateKey",
		Path: "/riki/deletePrivateKey/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorPrivateKey_c3k5r8"}`,
	},
	{
		Name: "SaveData",
		Path: "/riki/saveData/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorStoredData_4e18lv","dataID":"TestConnectorStoredData_0zeact","replaceCurrentItem":false,"data":"BNWozcdH3ul9qnwygeG0n1AOnTH8Tiobotu7WJx+oKcMLkPJkl1CWDKPo30HfyykGjbHWYkOVJFTsHy9xl09SoEatIxAgYa5YRJbKY8Ir5tqifFXT8bXRY4w6oK3catxN/TzK8WnPt3bDFs8o8A2jTA+Pxu3ngf9","plaintextHash":"c8416ccb379b9cc591a44b09765a806a4c164a1b23f287fb9846edf8760f3839"}`,
	},
	{
		Name: "GetData",
		Path: "/riki/getData/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorStoredData_4e18lv","dataID":"TestConnectorStoredData_0zeact"}`,
	},
	{
		Name: "DeleteData",
		Path: "/riki/deleteData/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorStoredData_4e18lv","dataID":"TestConnectorStoredData_0zeact"}`,
	},
	{
		Name: "DecryptData",
		Path: "/riki/decryptData/v1",
		Body: `{"apiKey":"/4pKidQqVz+SQ9G8G1pPJVb2","keyID":"TestConnectorStoredData_4e18lv","data":"BNWozcdH3ul9qnwygeG0n1AOnTH8Tiobotu7WJx+oKcMLkPJkl1CWDKPo30HfyykGjbHWYkOVJFTsHy9xl09SoEatIxAgYa5YRJbKY8Ir5tqifFXT8bXRY4w6oK3catxN/TzK8WnPt3bDFs8o8A2jTA+Pxu3ngf9"}`,
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
		App:    "riki",
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
