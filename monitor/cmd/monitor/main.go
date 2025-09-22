package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/monitor/internal"
)

func main() {
	var (
		ctx    = context.Background()
		err    error
		appEnv int
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
	redisDNS := os.Getenv("MONITOR_REDIS_DNS")
	if redisDNS == "" {
		log.Fatalf("must set env variable for 'MONITOR_REDIS_DNS'")
	}
	mongodbUri := os.Getenv("MONITOR_MONGODB_URI")
	if mongodbUri == "" {
		log.Fatalf("must set env variable for 'MONITOR_MONGODB_URI'")
	}
	diffCoinbaseBinanceCron := os.Getenv("DIFF_COINBASE_BINANCE_CRON")
	if diffCoinbaseBinanceCron == "" {
		log.Fatalf("must set env variable for 'DIFF_COINBASE_BINANCE_CRON'")
	}
	txsMempoolCron := os.Getenv("TXS_MEMPOOL_CRON")
	if txsMempoolCron == "" {
		log.Fatalf("must set env variable for 'TXS_MEMPOOL_CRON'")
	}
	_, err = internal.NewApplication(ctx, appName, appEnv, redisDNS, mongodbUri)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	cmd := <-c
	log.Printf("receive a signal: %v\n", cmd)
	log.Println("app exit")
}
