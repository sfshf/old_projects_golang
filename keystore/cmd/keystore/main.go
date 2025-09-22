package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nextsurfer/keystore/internal/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func ParseEnvFile(logger *zap.Logger) {
	envFile := "/etc/keystore/.env"
	f, err := os.Open(envFile)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if bytes.Equal(bytes.TrimSpace(line), []byte("")) {
			continue
		}
		kv := bytes.Split(line, []byte("="))
		if len(kv) == 2 {
			key := string(bytes.TrimSpace(kv[0]))
			value := string(bytes.TrimSpace(kv[1]))
			logger.Info("env variable", zap.String(key, value))
			if err := os.Setenv(key, value); err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	var err error
	// load zap logger
	cores := make([]zapcore.Core, 0)
	fw := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "/tmp/keystore.log",
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     3,
	})
	fw = zap.CombineWriteSyncers(fw, os.Stdout)
	je := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	cores = append(cores, zapcore.NewCore(je, fw, zap.NewAtomicLevel()))
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)

	// parse env file
	ParseEnvFile(logger)

	appEnvStr := os.Getenv("APPLICATION_ENV")
	if appEnvStr == "" {
		logger.Fatal("must set env variable for 'APPLICATION_ENV'")
	}
	appEnv, err := strconv.Atoi(appEnvStr)
	if err != nil {
		logger.Fatal("APPLICATION_ENV string is error", zap.NamedError("appError", err))
	}
	appHost := os.Getenv("KEYSTORE_SERVER_HOST")
	if appHost == "" {
		logger.Fatal("must set env variable for 'HTTP_SERVER_HOST'")
	}
	portStr := os.Getenv("KEYSTORE_SERVER_PORT")
	if portStr == "" {
		logger.Fatal("must set env variable for 'KEYSTORE_SERVER_PORT'")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Fatal("KEYSTORE_SERVER_PORT string is error", zap.NamedError("appError", err))
	}
	logger.Info("environment variables",
		zap.String("APPLICATION_ENV", appEnvStr),
		zap.String("KEYSTORE_SERVER_HOST", appHost),
		zap.Int("KEYSTORE_SERVER_PORT", port),
	)

	httpApp := app.NewKeyStoreApp("keystore", logger, port, appHost, appEnv)
	httpApp.Start()

	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	cmd := <-c
	httpApp.Logger.Info(fmt.Sprintf("receive a signal: %v", cmd))
	httpApp.Stop()
	httpApp.Logger.Info("servers stop gracefully")
}
