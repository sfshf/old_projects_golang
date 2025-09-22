package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/slark/internal/app/http"
)

func main() {
	var (
		err error
		ctx = context.Background()
	)
	// check environment variables
	appEnvStr := os.Getenv("APPLICATION_ENV")
	if appEnvStr == "" {
		log.Fatalln("must set env variable for 'APPLICATION_ENV'")
	}
	appEnv, err := strconv.Atoi(appEnvStr)
	if err != nil {
		log.Fatalln(fmt.Sprintf("appEnvStr is error : %s", appEnvStr))
	}
	appHost := "0.0.0.0"
	portStr := os.Getenv("HTTP_SERVER_PORT")
	if portStr == "" {
		log.Fatalln("must set env variable for 'HTTP_SERVER_PORT'")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalln(fmt.Sprintf("httpPortStr is error : %v", port))
	}
	mysqlDNS := os.Getenv("SLARK_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalln("must set env variable for 'SLARK_MYSQL_DNS'")
	}
	redisDNS := os.Getenv("SLARK_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalln("must set env variable for 'SLARK_REDIS_DNS'")
	}
	webPath := os.Getenv("SLARK_WEB_PATH")
	if webPath == "" {
		log.Fatalln("must set env variable for 'SLARK_WEB_PATH'")
	}
	discourseConnectSecret := os.Getenv("DISCOURSE_CONNECT_SECRET")
	if discourseConnectSecret == "" {
		log.Fatalln("must set env variable for 'DISCOURSE_CONNECT_SECRET'")
	}
	discourseConnectSignInAddress := os.Getenv("DISCOURSE_CONNECT_SIGN_IN_ADDRESS")
	if discourseConnectSignInAddress == "" {
		log.Fatalln("must set env variable for 'DISCOURSE_CONNECT_SIGN_IN_ADDRESS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalln("must set env variable for 'I18N_TOML_PATH'")
	}
	// build application instance
	app, err := http.NewApplication("slark-http", port, appHost, appEnv, redisDNS, mysqlDNS, webPath, tomlPath)
	if err != nil {
		log.Fatalln(err)
	}
	if err := app.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	// monitor signal
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	cmd := <-c
	log.Printf("receive a signal: %v\n", cmd)
	if err := app.Stop(ctx); err != nil {
		log.Fatalln(err)
	}
	log.Println("servers stop gracefully")
}
