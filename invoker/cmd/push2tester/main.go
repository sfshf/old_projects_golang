package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/invoker/internal/common/simplehttp"
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
		Name: "GetSite",
		Path: "/invoker/site/getSite/v1",
		Body: `{ "name": "Test" }`,
	},
	{
		Name: "GetCategories",
		Path: "/invoker/site/getCategories/v1",
		Body: `{ "siteID": "" }`,
	},
	{
		Name: "GetPosts",
		Path: "/invoker/site/getPosts/v1",
		Body: `{ "siteID": "", "categoryID": "" }`,
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
		App:    "invoker",
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
