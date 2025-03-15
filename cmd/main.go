package main

import (
	"fmt"

	"github.com/minikeyvalue/src/config"
	"github.com/minikeyvalue/src/storage"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println(fmt.Errorf("main: failed create logger instance: %w", err))
		return
	}
	cfg := config.ParseCommandFlags()
	fmt.Println(cfg.PathToStorageFile)

	st, err := storage.NewWithLoadData(cfg.PathToStorageFile)
	if err != nil {
		log.Error("Failed load storage data", zap.Error(err))
		return
	}

	if err := st.Add("aasfdssfdsfasfas", "adfas"); err != nil {
		log.Error("Failed add data to storage", zap.Error(err))
	}
}
