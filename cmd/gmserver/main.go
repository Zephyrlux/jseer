package main

import (
	"os"

	"jseer/internal/config"
	"jseer/internal/gm"
	"jseer/internal/logging"
	"jseer/internal/storage"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load(config.ResolvePath("configs/config.yaml"))
	if err != nil {
		panic(err)
	}
	logger, err := logging.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}

	store, err := storage.NewStore(cfg.Database)
	if err != nil {
		logger.Error("store init failed", zap.Error(err))
		os.Exit(1)
	}
	defer store.Close()

	server := gm.NewServer(cfg.GM, store, logger)
	if err := server.Run(cfg.GM.Address); err != nil {
		logger.Error("gm server stopped", zap.Error(err))
		os.Exit(1)
	}
}
