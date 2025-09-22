package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/oracle/internal/console"
	"go.uber.org/zap"
)

func main() {
	var err error
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		log.Fatalf("must set env variable for 'CONSUL_HTTP_ADDR'")
	}
	appEnvStr := os.Getenv("APP_ENV")
	if appEnvStr == "" {
		log.Fatalf("must set env variable for 'APP_ENV'")
	}
	appEnv, err := strconv.Atoi(appEnvStr)
	if err != nil {
		log.Fatalf("appEnvStr is error : %s", appEnvStr)
	}
	appNameStr := os.Getenv("CONSOLE_APP_NAME")
	if appNameStr == "" {
		log.Fatalf("must set env variable for 'CONSOLE_APP_NAME'")
	}
	consoleHost := os.Getenv("CONSOLE_HOST")
	if consoleHost == "" {
		log.Fatalf("must set env variable for 'CONSOLE_HOST'")
	}
	httpPortStr := os.Getenv("CONSOLE_HTTP_PORT")
	if httpPortStr == "" {
		log.Fatalf("must set env variable for 'CONSOLE_HTTP_PORT'")
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		log.Fatalf("httpPortStr is error : %v", httpPort)
	}
	grpcPortStr := os.Getenv("CONSOLE_GRPC_PORT")
	if grpcPortStr == "" {
		log.Fatalf("must set env variable for 'CONSOLE_GRPC_PORT'")
	}
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		log.Fatalf("grpcPortStr is error : %v", grpcPort)
	}
	mysqlDNS := os.Getenv("CONSOLE_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'CONSOLE_MYSQL_DNS'")
	}
	redisDNS := os.Getenv("CONSOLE_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'CONSOLE_REDIS_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	consoleWebPath := os.Getenv("CONSOLE_WEB_PATH")
	if consoleWebPath == "" {
		log.Fatalf("must set env variable for 'CONSOLE_WEB_PATH'")
	}
	protoStatisticCronSpec := os.Getenv("PROTO_STATISTIC_CRON_SPEC")
	if protoStatisticCronSpec == "" {
		log.Fatalf("must set env variable for 'PROTO_STATISTIC_CRON_SPEC'")
	}
	serviceHealthCheckCronSpec := os.Getenv("SERVICE_HEALTH_CHECK_CRON_SPEC")
	if serviceHealthCheckCronSpec == "" {
		log.Fatalf("must set env variable for 'SERVICE_HEALTH_CHECK_CRON_SPEC'")
	}
	gatewayApiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
	if gatewayApiHostname == "" {
		log.Fatalf("must set env variable for 'GATEWAY_API_HOSTNAME'")
	}
	refreshTlsCertificateCronSpec := os.Getenv("REFRESH_TLS_CERTIFICATE_CRON_SPEC")
	if refreshTlsCertificateCronSpec == "" {
		log.Fatalf("must set env variable for 'REFRESH_TLS_CERTIFICATE_CRON_SPEC'")
	}
	ctx := context.Background()
	// new a http app named oracle-console
	consoleApp, err := console.NewConsoleApp(ctx, appEnv, appNameStr, consoleHost, httpPort, grpcPort, mysqlDNS, redisDNS, tomlPath, consoleWebPath)
	if err != nil {
		log.Fatalln(err)
	}
	consoleApp.Run(ctx)
	// wait signal
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	consoleApp.Logger.Info("Console app receive a signal", zap.String("signal", sig.String()))
	if err := consoleApp.Stop(ctx); err != nil {
		consoleApp.Logger.Error("Console app stop", zap.NamedError("appError", err))
	}
	consoleApp.Logger.Info("Console app stop gracefully")
}
