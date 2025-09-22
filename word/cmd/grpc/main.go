package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/word/internal/app/grpc"
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
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatalf("must set env variable for 'AWS_REGION'")
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
	redisDNS := os.Getenv("WORD_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'WORD_REDIS_DNS'")
	}
	mysqlDNS := os.Getenv("WORD_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'WORD_MYSQL_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	progressBackupBucketName := os.Getenv("PROGRESS_BACKUP_BUCKET_NAME")
	if progressBackupBucketName == "" {
		log.Fatalf("must set env variable for 'PROGRESS_BACKUP_BUCKET_NAME'")
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalf("must set env variable for 'API_KEY'")
	}
	studyBackupKey := os.Getenv("STUDY_BACKUP_KEY")
	if studyBackupKey == "" {
		log.Fatalf("must set env variable for 'STUDY_BACKUP_KEY'")
	}
	app, err := grpc.NewApplication(appName, grpcPort, appHost, appEnv, redisDNS, mysqlDNS, tomlPath)
	app.Logger.Info("Test Log TODO")
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
	log.Println("servers stop gracefully")
}
