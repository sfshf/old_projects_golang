package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nextsurfer/word/internal/pkg/simplehttp"
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
		Name: "FetchAudioURL",
		Path: "/word/audio/getAudioURL/v1",
		Body: `{
			"text":   "hello aws polly",
			"ssml":   "",
			"accent": "us",
			"voice":  "Matthew",
			"apiKey": "75df8c8c2b54d146305b01a9e0d649de",
		}`,
	},
	{
		Name: "FavoriteDefinition",
		Path: "/word/user/definition/favorite/v1",
		Body: `{
			"definitionID": 199999999,
		}`,
	},
	{
		Name: "FavoritedDefinitions",
		Path: "/word/user/definition/favorites/v1",
		Body: "",
	},
	{
		Name: "ProgressBackupStatus",
		Path: "/word/user/progress/backup/status/v1",
		Body: `{
			"timestamp": "now()",
			"version":   1,
		}`,
	},
	{
		Name: "UploadProgressBackup",
		Path: "/word/user/progress/backup/upload/v1",
		Body: `{
			"timestamp": "now()",
			"version":   1,
			"data":      "word e2e testing TestUploadProgressBackup.",
		}`,
	},
	{
		Name: "DownloadProgressBackup",
		Path: "/word/user/progress/backup/download/v1",
		Body: `{
			"timestamp": "now()",
			"version":   1,
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
		App:    "word",
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
