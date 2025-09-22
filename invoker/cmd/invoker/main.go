package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/invoker/internal"
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
	redisDNS := os.Getenv("INVOKER_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'INVOKER_REDIS_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	mysqlDNS := os.Getenv("INVOKER_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'INVOKER_MYSQL_DNS'")
	}
	app, err := internal.NewApplication(ctx, appName, grpcPort, appHost, appEnv, redisDNS, tomlPath, mysqlDNS)
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
