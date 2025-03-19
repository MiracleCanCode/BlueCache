package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/minikeyvalue/src/config"
	"github.com/minikeyvalue/src/storage"
	"github.com/minikeyvalue/src/transport/tcp"
	"github.com/minikeyvalue/src/transport/tcp/handlers"
	"github.com/minikeyvalue/src/utils/aof"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println(fmt.Errorf("main: failed create logger instance: %w", err))
		return
	}
	cfg := config.ParseCommandFlags()
	aofManager := aof.New(cfg.PathToStorageFile, log)
	storageInstance := storage.New(aofManager)
	transport, err := tcp.NewWithConn(cfg.Port)
	if err != nil {
		log.Error("Failed create tcp listener", zap.Error(err))
		return
	}

	log.Info("IsaRedis start work", zap.String("port", cfg.Port))

	var wg sync.WaitGroup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGBUS, syscall.SIGINT)
	if err := aofManager.RecoverData(cfg.Port); err != nil {
		log.Error("Failed recover data", zap.Error(err))
	}
	go func() {
		<-quit
		log.Info("IsaRedis shutting down.....", zap.Time("time", time.Now()))
		wg.Wait()
		transport.CloseConn()
		log.Info("IsaRedis stop work", zap.Time("isa_redis_stopped_time", time.Now()))
	}()

	for {
		conn, err := transport.Conn.Accept()
		if err != nil {
			log.Error("Failed listen new connections", zap.Error(err))
			return
		}
		log.Info("Client connected!")
		handler := handlers.NewStorage(log, conn, storageInstance)
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.HandleClient()
		}()
	}
}
