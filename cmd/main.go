package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/minikeyvalue/src/aof"
	"github.com/minikeyvalue/src/config"
	recoverdatafromaof "github.com/minikeyvalue/src/recoverDataFromAof"
	"github.com/minikeyvalue/src/storage"
	"github.com/minikeyvalue/src/transport/tcp"
	"github.com/minikeyvalue/src/transport/tcp/handlers"
	"github.com/minikeyvalue/src/utils/retry"
	"github.com/minikeyvalue/src/utils/timeout"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println(fmt.Errorf("main: failed create logger instance: %w", err))
		return
	}
	defer log.Sync()
	cfg, err := config.New()
	if err != nil {
		log.Error("Config error", zap.Error(err))
	}
	aofManager, err := aof.NewAOF(cfg.PathToStorageFile, log)
	if err != nil {
		log.Error("Failed create aof manager instance", zap.Error(err))
		return
	}
	storageInstance := storage.New(aofManager, false)
	recoveryStorage := storage.New(aofManager, true)
	recoverData := recoverdatafromaof.New(recoveryStorage)
	baseRetryDelayMiliseconds := 300
	retryAttempts := 5
	timeoutMillisecondsForRecoverData := 600

	recoveryDataFn := func() error {
		if err := recoverData.Recover(cfg.PathToStorageFile); err != nil {
			return fmt.Errorf("RecoverData: failed recover data: %w", err)
		}
		return nil
	}

	if err := timeout.Operation(timeoutMillisecondsForRecoverData,
		recoveryDataFn); err != nil {
		log.Error("Timeout for recover data operation!", zap.Error(err))
		return
	}

	var transport *tcp.TcpConn
	createTCPConnection := func() error {
		transport, err = tcp.NewWithConn(cfg.Port)
		if err != nil {
			return fmt.Errorf("CreateTcpConnection: failed create tcp connection: %w", err)
		}
		return nil
	}
	if err := retry.RetryOperation(log, createTCPConnection,
		baseRetryDelayMiliseconds,
		retryAttempts); err != nil {
		log.Error("Failed create tcp connection", zap.Error(err))
		return
	}

	log.Info("IsaRedis start work", zap.String("port", cfg.Port), zap.Bool("logging", cfg.Logging),
		zap.String("storage_path", cfg.PathToStorageFile))

	var wg sync.WaitGroup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGBUS)

	go func() {
		<-quit
		log.Info("IsaRedis shutting down.....", zap.Time("time", time.Now()))
		transport.CloseConn()
		wg.Wait()
		log.Info("IsaRedis stop work", zap.Time("isa_redis_stopped_time", time.Now()))
	}()

	for {
		conn, err := transport.Conn.Accept()
		if err != nil {
			log.Error("Failed listen new connections", zap.Error(err))
			break
		}
		log.Info("Client connected!")

		handler := handlers.NewStorageHandler(log, conn, storageInstance, cfg)
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.HandleClient()
		}()
	}
}
