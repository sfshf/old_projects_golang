package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/slark/internal/pkg/simplehttp"
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
		Name: "LogOutBySession",
		Path: "/slark/user/logout/v1",
		Body: ``,
	},
	{
		Name: "LoginByApple",
		Path: "/slark/user/loginByApple/v1",
		Body: `{"email":"TestLoginByApple@TestLoginByApple.net","userIdentifier":"TestLoginByApple"}`,
	},
	{
		Name: "LoginByEmailCode",
		Path: "/slark/loginByEmailCode/v1",
		Body: `{"email":"email@TestLoginByEmailCode.net","code":"100360"}`,
	},
	{
		Name: "LoginByEmail",
		Path: "/slark/user/loginByEmail/v1",
		Body: `{"email":"gavin@n1xt.net","passwordHash":"6af794e1f030fb6f8bb48c7be9620d59311fca1d0ebfb682bad0c0431d0f3ad7"}`,
	},
	{
		Name: "RandomNickname",
		Path: "/slark/randomNickname/v1",
		Body: ``,
	},
	{
		Name: "RegisterByEmail",
		Path: "/slark/user/registerByEmail/v1",
		Body: `{"email":"TestRegisterByEmail@n1xt.net","nickname":"nickname-TestRegisterByEmail","passwordHash":"99a588b044bf1b4435d1d29524b6eb79bacd8a89ef63365c56eae06c01bb25bd","captcha":"637999"}`,
	},
	{
		Name: "SendLoginEmailCode",
		Path: "/slark/sendLoginEmailCode/v1",
		Body: `{"email":"e2e-TestSendLoginEmailCode@n1xt.net"}`,
	},
	{
		Name: "SendRegistrationEmailCaptcha",
		Path: "/slark/user/sendRegistrationEmailCaptcha/v1",
		Body: `{"email":"e2e-TestSendRegistrationEmailCaptcha@n1xt.net"}`,
	},
	{
		Name: "SendRegistrationEmailCaptcha",
		Path: "/slark/user/sendRegistrationEmailCaptcha/v1",
		Body: `{"email":"e2e-TestSendRegistrationEmailCaptchaToRegisteredEmail@n1xt.net"}`,
	},
	{
		Name: "Unregister",
		Path: "/slark/unregister/v1",
		Body: ``,
	},
	{
		Name: "UpdateNickname",
		Path: "/slark/updateNickname/v1",
		Body: `{"nickname":"TestUpdateNickname199509"}`,
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
		App:    "slark",
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
