package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"jseer/internal/config"
	"jseer/internal/loginserver"
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

	srv := loginserver.New(cfg.Login, cfg.Game, store, logger)
	loginserver.RegisterHandlers(srv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Start(ctx); err != nil {
		logger.Error("login server stopped", zap.Error(err))
		os.Exit(1)
	}
}
