package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	oracle_console_api "github.com/nextsurfer/oracle/api/console"
	oracle_console_grpc "github.com/nextsurfer/oracle/pkg/console/grpc"
)

type ServiceConfig struct {
	Name          string
	Application   string
	Url           string
	PathPrefix    string
	ProtoFilePath string
	ProtoFile     string
}

var services = []ServiceConfig{
	{
		Name:          "doom",
		Application:   "doom",
		Url:           "",
		PathPrefix:    "/doom",
		ProtoFilePath: "api/http/doom-go.http.proto",
	},
}

func PushService(config ServiceConfig) error {
	ctx := context.Background()
	respData, err := oracle_console_grpc.UpsertService(ctx, &oracle_console_api.UpsertServiceRequest{
		Name:        config.Name,
		Application: config.Application,
		Url:         config.Url,
		PathPrefix:  config.PathPrefix,
		ProtoFile:   config.ProtoFile,
	})
	if err != nil {
		return err
	}
	if respData.Code != 0 {
		return fmt.Errorf("PushService: %v", respData)
	}
	return nil
}

func main() {
	for _, service := range services {
		f, err := os.Open(service.ProtoFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		data, err := io.ReadAll(f)
		if err != nil {
			log.Fatalln(err)
		}
		service.ProtoFile = base64.StdEncoding.EncodeToString(data)
		if err := PushService(service); err != nil {
			log.Fatalln(err)
		}
	}

}
