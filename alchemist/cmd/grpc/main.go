package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/alchemist/internal/app/grpc"
	"github.com/rs/cors"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	var (
		err      error
		grpcPort int
		appEnv   int
	)
	// check environment variables
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		log.Fatalf("must set env variable for 'APP_NAME'")
	}
	connectorKeyID := os.Getenv("CONNECTOR_KEY_ID")
	if connectorKeyID == "" {
		log.Fatalf("must set env variable for 'CONNECTOR_KEY_ID'")
	}
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		log.Fatalf("must set env variable for 'CONSUL_HTTP_ADDR'")
	}
	appHost := os.Getenv("GRPC_SERVER_HOST")
	if appHost == "" {
		log.Fatalf("must set env variable for 'GRPC_SERVER_ADDR'")
	}
	grpcPortStr := os.Getenv("GRPC_SERVER_PORT")
	if grpcPortStr == "" {
		log.Fatalf("must set env variable for 'GRPC_SERVER_PORT'")
	}
	grpcPort, err = strconv.Atoi(grpcPortStr)
	if err != nil {
		log.Fatalf("grpcPortStr is error : %s", grpcPortStr)
	}
	httpPortStr := os.Getenv("HTTP_SERVER_PORT")
	if httpPortStr == "" {
		log.Fatalf("must set env variable for 'HTTP_SERVER_PORT'")
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		log.Fatalf("httpPortStr is error : %s", httpPortStr)
	}
	managerWebPath := os.Getenv("ALCHEMIST_MANAGER_WEB_PATH")
	if managerWebPath == "" {
		log.Fatalf("must set env variable for 'ALCHEMIST_MANAGER_WEB_PATH'")
	}
	appEnvStr := os.Getenv("APPLICATION_ENV")
	if appEnvStr == "" {
		log.Fatalf("must set env variable for 'APPLICATION_ENV'")
	}
	appEnv, err = strconv.Atoi(appEnvStr)
	if err != nil {
		log.Fatalf("appEnvStr is error : %s", appEnvStr)
	}
	redisDNS := os.Getenv("ALCHEMIST_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'ALCHEMIST_REDIS_DNS'")
	}
	mysqlDNS := os.Getenv("ALCHEMIST_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'ALCHEMIST_MYSQL_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	// http web server
	http.Handle("/", http.FileServer(http.Dir(managerWebPath)))
	// cors
	handler := cors.Default().Handler(http.DefaultServeMux)
	addr := fmt.Sprintf("%s:%v", "0.0.0.0", httpPort)
	server := &http.Server{Addr: addr, Handler: handler}
	go func() {
		log.Printf("HTTP api server is starting at %s ...\n", addr)
		log.Fatal(server.ListenAndServe())
	}()
	// grpc server
	app, err := grpc.NewApplication(appName, grpcPort, appHost, appEnv, redisDNS, mysqlDNS, tomlPath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := app.Start(); err != nil {
		log.Fatalln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	cmd := <-c
	log.Printf("receive a signal: %v\n", cmd)
	if err := app.Stop(); err != nil {
		log.Fatalln(err)
	}
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalln(err)
	}
	log.Println("servers stop gracefully")
}
