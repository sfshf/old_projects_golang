package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/nextsurfer/ground/pkg/util"
	gutil "github.com/nextsurfer/ground/pkg/util"
	"github.com/nextsurfer/slark/internal/app/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	var (
		err      error
		grpcPort int
		appEnv   int
		ctx      = context.Background()
	)
	// check environment variables
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		log.Fatalf("must set env variable for 'APP_NAME'")
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
	appEnvStr := os.Getenv("APPLICATION_ENV")
	if appEnvStr == "" {
		log.Fatalf("must set env variable for 'APPLICATION_ENV'")
	}
	appEnv, err = strconv.Atoi(appEnvStr)
	if err != nil {
		log.Fatalf("appEnvStr is error : %s", appEnvStr)
	}
	redisDNS := os.Getenv("SLARK_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'SLARK_REDIS_DNS'")
	}
	mysqlDNS := os.Getenv("SLARK_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'SLARK_MYSQL_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	var mongodbUri string
	if gutil.EnvForInt(appEnv) == gutil.AppEnvDEV || gutil.EnvForInt(appEnv) == gutil.AppEnvPPE {
		mongodbUri = os.Getenv("SLARK_MONGODB_URI")
		if mongodbUri == "" {
			log.Fatalf("must set env variable for 'SLARK_MONGODB_URI'")
		}
	}
	// build application instance
	app, err := grpc.NewApplication(ctx, appName, grpcPort, appHost, appEnv, redisDNS, mysqlDNS, tomlPath, mongodbUri)
	if err != nil {
		log.Fatalln(err)
	}
	if err := app.Start(); err != nil {
		log.Fatalln(err)
	}
	// monitor signal
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
