package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/doom-go/internal"
)

func main() {
	var (
		ctx      = context.Background()
		err      error
		grpcPort int
		appEnv   int
	)
	// check environment variables
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		log.Fatalf("must set env variable for 'APP_NAME'")
	}
	appEnvStr := os.Getenv("APPLICATION_ENV")
	if appEnvStr == "" {
		log.Fatalf("must set env variable for 'APPLICATION_ENV'")
	}
	appEnv, err = strconv.Atoi(appEnvStr)
	if err != nil {
		log.Fatalf("appEnvStr is error : %s", appEnvStr)
	}
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		log.Fatalf("must set env variable for 'CONSUL_HTTP_ADDR'")
	}
	appHost := os.Getenv("GRPC_SERVER_HOST")
	if appHost == "" {
		log.Fatalf("must set env variable for 'GRPC_SERVER_HOST'")
	}
	grpcPortStr := os.Getenv("GRPC_SERVER_PORT")
	if grpcPortStr == "" {
		log.Fatalf("must set env variable for 'GRPC_SERVER_PORT'")
	}
	grpcPort, err = strconv.Atoi(grpcPortStr)
	if err != nil {
		log.Fatalf(fmt.Sprintf("grpcPortStr is error : %s", grpcPortStr))
	}
	redisDNS := os.Getenv("DOOM_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'DOOM_REDIS_DNS'")
	}
	connectorApiKey := os.Getenv("CONNECTOR_API_KEY")
	if connectorApiKey == "" {
		log.Fatalf("must set env variable for 'CONNECTOR_API_KEY'")
	}
	connectorKeyID := os.Getenv("CONNECTOR_KEY_ID")
	if connectorKeyID == "" {
		log.Fatalf("must set env variable for 'CONNECTOR_KEY_ID'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	if os.Getenv("CRYPTOCURRENCY_CHAIN_CONFIG_PATH") == "" {
		log.Fatalf("must set env variable for 'CRYPTOCURRENCY_CHAIN_CONFIG_PATH'")
	}
	mongodbUri := os.Getenv("DOOM_MONGODB_URI")
	if mongodbUri == "" {
		log.Fatalf("must set env variable for 'DOOM_MONGODB_URI'")
	}
	app, err := internal.NewApplication(ctx, appName, connectorApiKey, connectorKeyID, grpcPort, appHost, appEnv, redisDNS, tomlPath, mongodbUri)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if err := app.Start(); err != nil {
		log.Fatalf(err.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	cmd := <-c
	log.Printf("receive a signal: %v\n", cmd)
	if err := app.Stop(); err != nil {
		log.Fatalln(err)
	}
	log.Println("servers stop gracefully")
}
