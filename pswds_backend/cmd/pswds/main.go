package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/pswds_backend/internal"
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
	redisDNS := os.Getenv("PSWDS_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'PSWDS_REDIS_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	mysqlDNS := os.Getenv("PSWDS_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'PSWDS_MYSQL_DNS'")
	}
	privacyEmailAdminAccount := os.Getenv("PRIVACY_EMAIL_ADMIN_ACCOUNT")
	if privacyEmailAdminAccount == "" {
		log.Fatalf("must set env variable for 'PRIVACY_EMAIL_ADMIN_ACCOUNT'")
	}
	privacyEmailAdminPassword := os.Getenv("PRIVACY_EMAIL_ADMIN_PASSWORD")
	if privacyEmailAdminPassword == "" {
		log.Fatalf("must set env variable for 'PRIVACY_EMAIL_ADMIN_PASSWORD'")
	}
	emailServerHost := os.Getenv("EMAIL_SERVER_HOST")
	if emailServerHost == "" {
		log.Fatalf("must set env variable for 'EMAIL_SERVER_HOST'")
	}
	emailServerPort := os.Getenv("EMAIL_SERVER_PORT")
	if emailServerPort == "" {
		log.Fatalf("must set env variable for 'EMAIL_SERVER_PORT'")
	}
	emailServerUsername := os.Getenv("EMAIL_SERVER_USERNAME")
	if emailServerUsername == "" {
		log.Fatalf("must set env variable for 'EMAIL_SERVER_USERNAME'")
	}
	emailServerPassword := os.Getenv("EMAIL_SERVER_PASSWORD")
	if emailServerPassword == "" {
		log.Fatalf("must set env variable for 'EMAIL_SERVER_PASSWORD'")
	}
	emailServerFrom := os.Getenv("EMAIL_SERVER_FROM")
	if emailServerFrom == "" {
		log.Fatalf("must set env variable for 'EMAIL_SERVER_FROM'")
	}
	// 密码找回限流：以24小时为单位
	recoverLimitPeriod := os.Getenv("RECOVER_LIMIT_PERIOD")
	if recoverLimitPeriod == "" {
		log.Fatalf("must set env variable for 'RECOVER_LIMIT_PERIOD'")
	}
	if _, err := strconv.Atoi(recoverLimitPeriod); err != nil {
		log.Fatalf("invalid value for 'RECOVER_LIMIT_PERIOD'")
	}
	// 密码家庭找回拒绝期：以24小时为单位
	familyRecoverProbationPeriod := os.Getenv("FAMILY_RECOVER_PROBATION_PERIOD")
	if familyRecoverProbationPeriod == "" {
		log.Fatalf("must set env variable for 'FAMILY_RECOVER_PROBATION_PERIOD'")
	}
	if _, err := strconv.Atoi(familyRecoverProbationPeriod); err != nil {
		log.Fatalf("invalid value for 'FAMILY_RECOVER_PROBATION_PERIOD'")
	}
	app, err := internal.NewApplication(ctx, appName, grpcPort, appHost, appEnv, redisDNS, tomlPath, mysqlDNS)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if err := app.Start(); err != nil {
		log.Fatalln(err.Error())
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
