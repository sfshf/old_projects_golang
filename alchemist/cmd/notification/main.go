package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	alchemist_notification_api "github.com/nextsurfer/alchemist/api/notification"
	"github.com/nextsurfer/alchemist/internal/app/notification"
	grpcservers "github.com/nextsurfer/alchemist/internal/app/notification/servers"
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
	redisDNS := os.Getenv("ALCHEMIST_NOTIFICATION_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'ALCHEMIST_NOTIFICATION_REDIS_DNS'")
	}
	mysqlDNS := os.Getenv("ALCHEMIST_NOTIFICATION_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'ALCHEMIST_NOTIFICATION_MYSQL_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	app, err := notification.NewApplication(appName, grpcPort, appHost, appEnv, redisDNS, mysqlDNS, tomlPath)
	if err != nil {
		log.Fatalln(err)
	}
	// register servers
	alchemist_notification_api.RegisterAlchemistNotificationServiceServer(app.Server.GrpcServer(), grpcservers.NewAlchemistNotificationServer(app.Env, app.Logger, app.DaoManager, app.RedisOption, app.Server.LocolizerManager, app.Validator))
	app.Logger.Info("RegisterAlchemistNotificationServiceServer success")

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
	log.Println("servers stop gracefully")
}
