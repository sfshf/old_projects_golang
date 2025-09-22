package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
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
		Name: "CreateSecurityQuestions",
		Path: "/doom/createSecurityQuestions/v1",
		Body: `{
		  "plainText":
			"VGVzdENyZWF0ZVNlY3VyaXR5UXVlc3Rpb25zLXBsYWludGV4dF9ya2Eya3A1ZHBkYnBrbjVi",
		  "cipherText":
			"BEFuG5YbjMVUf42Vp7Z/JFfMsu3OyIWe67Uu5CpwhtYmx1idbFA9H+TiWghtOm+LvvPGTyu00aqQRstW5V9yzTN+NKkc5PKJPbX1mYxZsKo3IP4xjotI+9I4/2qXerqJwQ9PV51jLRK3bDO24yXPavXG5GZQ1t63i54naHOTZ0KC2bgETtm5o8qReyATJYsRn7rancJN91nhdvg30lc172diBTwkejMoMg==",
		  "title": "TestCreateSecurityQuestions-title_hl7n0qrfo03uopz7",
		  "description": "TestCreateSecurityQuestions-description_kd2bamv1gdhue2il",
		}`,
	},
	{
		Name: "ListSecurityQuestions",
		Path: "/doom/listSecurityQuestions/v1",
		Body: ``,
	},
	{
		Name: "GetSecurityQuestions",
		Path: "/doom/getSecurityQuestions/v1",
		Body: `{ "title": "TestGetSecurityQuestions-title_frmnv1q9qsbd5yly" }`,
	},
	{
		Name: "DeleteSecurityQuestions",
		Path: "/doom/deleteSecurityQuestions/v1",
		Body: `{ "title": "TestDeleteSecurityQuestions-title_1cf7om0cnjobkf0d" }`,
	},
	{
		Name: "CreateData",
		Path: "/doom/createData/v1",
		Body: `{
		  "plainText":
			"eyJxdWVzdGlvbjEiOiJhc2RmIiwicXVlc3Rpb24yIjoiYXNkZiIsInF1ZXN0aW9uMyI6ImFzZGYiLCJlbmNyeXB0ZWRQYXNzd29yZCI6IlgyQTF2UUdIOElIMHI5Ry9yVU5kbXcrNmRtcz0iLCJoYXNoT2ZIYXNoIjoiMzdmZmRjN2E4YzJiYzc2NWE4NWIxYzQxOTY1ZjgxZjg1OWYzNTJkMjlkM2E2ZDlmNmVhNTM1M2UyNWZjYzMzNiIsIm5vbmNlIjoiKzNVVkkzcXFhWmV0M0FTM1pKY3JsQUZDK2F0V2ltSVMifQ==",
		  "cipherText":
			"BNvA7AJFaBIiBc0gGgAHUUb2LL82wynqZgm+pDELjvY+G9xDPcKxu0vBZlsFaKLoKJVcg2rEw3eNVfL9Gllxyti0ZbznmLB6d/bB8WNkgIBWpvVFYE5HAMkAKPfzIAVS7HphR7g3SE8p2vW1M8PF0FFX5OjDPraEogYUk2wnzSVAXpPf3r5g9nSgC3eYrbI71RRUXOkd1WwuAFF+ReCUgONwTyvt1PvqaXrG2Zbv4dRNHR9pxg5HfSK+Js8s2ZPbo2alVPXamcUkeha5oXY4dojwU5GrvBMRGELVyFYd/dVTJ82N7LjfWYb+rUP86SVvHYrz+BLh18n6yKMPfHVZEefA+h3+gfdnwWPTgxqx5U72/mfES2oUEzkD7WGyMiRjXn6Ys17iZBHrxadq6xAEpuRje+OBCHaxmITVFhsT34QwOkOnHuJjfvbwUhJCMzCWNCHuknIjhTUI4rHEHltj/MNZxZIoi93dWVHeqgkJLyvUQe8DKz2pcH7tBC5HdPLarVC4sI62hSeCE1PR8PxzGq0NGQ3QjsC5yw==",
		  "title": "TestJsEncrypt",
		}`,
	},
	{
		Name: "ListData",
		Path: "/doom/listData/v1",
		Body: ``,
	},
	{
		Name: "GetData",
		Path: "/doom/getData/v1",
		Body: `{ "title": "TestJsEncrypt" }`,
	},
	{
		Name: "DeleteData",
		Path: "/doom/deleteData/v1",
		Body: `{ "title": "TestJsEncrypt" }`,
	},
	{
		Name: "GetLatestSpotPrice",
		Path: "/doom/getLatestSpotPrice/v1",
		Body: `{ "symbol": "CRO", baseCoin: "USDT" }`,
	},
	{
		Name: "GetLatestSpotPrices",
		Path: "/doom/getLatestSpotPrices/v1",
		Body: `{ "symbols": ["CRO", "weth", "eth", "OKB"], baseCoin: "USDT" }`,
	},
	{
		Name: "GetAssets",
		Path: "/doom/getAssets/v1",
		Body: `{
		  "address": "3DdfA8eC3052539b6C9549F12cEA2C295cfF5296",
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetUIKlines",
		Path: "/doom/getUIKlines/v1",
		Body: `{
		  "beginTime": "TIMESTAMP=n-24h",
		  "endTime": "TIMESTAMP=n",
		  "baseCoin": "USDT",
		  "symbol": "usdc",
		  "interval": "1h",
		}`,
	},
	{
		Name: "GetDappAssets",
		Path: "/doom/getDappAssets/v1",
		Body: `{
		  "address": "3DdfA8eC3052539b6C9549F12cEA2C295cfF5296",
		  "chain": "eth",
		  "app": "aave_v3",
		}`,
	},
	{
		Name: "GetTokens",
		Path: "/doom/getTokens/v1",
		Body: ``,
	},
	{
		Name: "GetDapps",
		Path: "/doom/getDapps/v1",
		Body: ``,
	},
	{
		Name: "FavoriteToken",
		Path: "/doom/favoriteToken/v1",
		Body: `{ "symbol": "weth" }`,
	},
	{
		Name: "GetFavoritedTokens",
		Path: "/doom/getFavoritedTokens/v1",
		Body: ``,
	},
	{
		Name: "GetFavoritedLatestSpotPrices",
		Path: "/doom/getFavoritedLatestSpotPrices/v1",
		Body: `{ "symbols": ["eth"], baseCoin: "USDT" }`,
	},
	{
		Name: "GetTokenBalances",
		Path: "/doom/getTokenBalances/v1",
		Body: `{
		  "address": "36cc7B13029B5DEe4034745FB4F24034f3F2ffc6",
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetBalance",
		Path: "/doom/getBalance/v1",
		Body: `{
		  "address": "36cc7B13029B5DEe4034745FB4F24034f3F2ffc6",
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetABI",
		Path: "/doom/getABI/v1",
		Body: `{
		  "address": "A0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetTokenApprovals",
		Path: "/doom/getTokenApprovals/v1",
		Body: `{
		  "address": "36cc7B13029B5DEe4034745FB4F24034f3F2ffc6",
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetAddress",
		Path: "/doom/getAddress/v1",
		Body: `{
		  "address": "bc1qzfnufa2v2rkaaaqapwmsyxs7vq2fu9fq4h9n7h",
		}`,
	},
	{
		Name: "GetAddressTransactions",
		Path: "/doom/getAddressTransactions/v1",
		Body: `{
		  "address": "bc1qzfnufa2v2rkaaaqapwmsyxs7vq2fu9fq4h9n7h",
		}`,
	},
	{
		Name: "GetAddressTransactionsMempool",
		Path: "/doom/getAddressTransactionsMempool/v1",
		Body: `{
		  "address": "bc1qzfnufa2v2rkaaaqapwmsyxs7vq2fu9fq4h9n7h",
		}`,
	},
	{
		Name: "GetGasFee",
		Path: "/doom/getGasFee/v1",
		Body: `{
		  "chain": "eth",
		}`,
	},
	{
		Name: "GetEstimationOfConfirmationTime",
		Path: "/doom/getEstimationOfConfirmationTime/v1",
		Body: `{
		  "chain": "eth",
		  "gasPrice": "200000"
		}`,
	},
	{
		Name: "CreatePrivateKeyBackup",
		Path: "/doom/createPrivateKeyBackup/v1",
		Body: `{
		  "plainText":
			"VGVzdENyZWF0ZVByaXZhdGVLZXlCYWNrdXAtcGxhaW50ZXh0XzFsZGd2cmV6dGFvZXBnYWw=",
		  "cipherText":
			"BPGeKcuemKsPh7Fp4LdwxFBgOXD/c8YxMNrxALW8S6ku/gKMCgYlRTNjYmselQQPtjqobDJdvdz4ZU6XMdC0woeegQFQ6HSREjXlWhmEF/h1AuqQPXIzMZbWhGfFuduDrYZ6Ji24Lwfbx2izGm4uON4NAVpWm3Bkt6zXL0iV6ST3XKS7JyFu+cSZ0ZJsAokw3Z9yGARWC3HfRCtck302A/QiStD+Npi21Q==",
		  "title": "TestCreatePrivateKeyBackup-title_p77knjwqzgjbfj7c",
		}`,
	},
	{
		Name: "ListPrivateKeyBackup",
		Path: "/doom/listPrivateKeyBackup/v1",
		Body: ``,
	},
	{
		Name: "GetPrivateKeyBackup",
		Path: "/doom/getPrivateKeyBackup/v1",
		Body: `{
		  "title": "TestGetPrivateKeyBackup-title_872cpm9yxa8jw8sw",
		}`,
	},
	{
		Name: "DeletePrivateKeyBackup",
		Path: "/doom/deletePrivateKeyBackup/v1",
		Body: `{
		  "title": "TestDeletePrivateKeyBackup-title_9qhyz0w76kt6tkml",
		}`,
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
		App:    "doom",
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
