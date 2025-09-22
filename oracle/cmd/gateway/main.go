package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/oracle/internal/gateway"
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
	appNameStr := os.Getenv("GATEWAY_APP_NAME")
	if appNameStr == "" {
		log.Fatalf("must set env variable for 'GATEWAY_APP_NAME'")
	}
	gatewayHost := os.Getenv("GATEWAY_HOST")
	if gatewayHost == "" {
		log.Fatalf("must set env variable for 'GATEWAY_HOST'")
	}
	httpPortStr := os.Getenv("GATEWAY_HTTP_PORT")
	if httpPortStr == "" {
		log.Fatalf("must set env variable for 'GATEWAY_HTTP_PORT'")
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		log.Fatalf("httpPortStr is error : %v", httpPort)
	}
	tlsPortStr := os.Getenv("GATEWAY_TLS_PORT")
	if tlsPortStr == "" {
		log.Fatalf("must set env variable for 'GATEWAY_TLS_PORT'")
	}
	tlsPort, err := strconv.Atoi(tlsPortStr)
	if err != nil {
		log.Fatalf("tlsPortStr is error : %v", tlsPort)
	}
	rpcPortStr := os.Getenv("GATEWAY_GRPC_PORT")
	if rpcPortStr == "" {
		log.Fatalf("must set env variable for 'GATEWAY_GRPC_PORT'")
	}
	rpcPort, err := strconv.Atoi(rpcPortStr)
	if err != nil {
		log.Fatalf("rpcPortStr is error : %v", rpcPort)
	}
	mysqlDNS := os.Getenv("GATEWAY_MYSQL_DNS")
	if mysqlDNS == "" {
		log.Fatalf("must set env variable for 'GATEWAY_MYSQL_DNS'")
	}
	redisDNS := os.Getenv("GATEWAY_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'GATEWAY_REDIS_DNS'")
	}
	tomlPath := os.Getenv("I18N_TOML_PATH")
	if tomlPath == "" {
		log.Fatalf("must set env variable for 'I18N_TOML_PATH'")
	}
	consoleHost := os.Getenv("CONSOLE_HOST")
	if consoleHost == "" {
		log.Fatalf("must set env variable for 'CONSOLE_HOST'")
	}
	consoleGrpcPortStr := os.Getenv("CONSOLE_GRPC_PORT")
	if consoleGrpcPortStr == "" {
		log.Fatalf("must set env variable for 'CONSOLE_GRPC_PORT'")
	}
	gatewayApiHostname := os.Getenv("GATEWAY_API_HOSTNAME")
	if gatewayApiHostname == "" {
		log.Fatalf("must set env variable for 'GATEWAY_API_HOSTNAME'")
	}
	uploadBucketName := os.Getenv("UPLOAD_BUCKET_NAME")
	if uploadBucketName == "" {
		log.Fatalf("must set env variable for 'UPLOAD_BUCKET_NAME'")
	}
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatalf("must set env variable for 'AWS_REGION'")
	}
	uploadProtoStatisticCronSpec := os.Getenv("UPLOAD_PROTO_STATISTIC_CRON_SPEC")
	if uploadProtoStatisticCronSpec == "" {
		log.Fatalf("must set env variable for 'UPLOAD_PROTO_STATISTIC_CRON_SPEC'")
	}
	gatewayInitCorsOrigins := os.Getenv("ORACLE_GATEWAY_INIT_CORS")
	if gatewayInitCorsOrigins == "" {
		log.Fatalf("must set env variable for 'ORACLE_GATEWAY_INIT_CORS'")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// new a http app named oracle-admin
	gatewayApp, err := gateway.NewGatewayApp(ctx, appEnv, appNameStr, gatewayHost, httpPort, tlsPort, rpcPort, mysqlDNS, redisDNS, tomlPath)
	if err != nil {
		log.Fatalln(err)
	}
	if err := gatewayApp.Run(ctx); err != nil {
		log.Fatalln(err)
	}
	// wait signal
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	gatewayApp.Logger.Info("Gateway app receive a signal", zap.String("signal", sig.String()))
	if err := gatewayApp.Stop(ctx); err != nil {
		log.Fatalln(err)
	}
}
