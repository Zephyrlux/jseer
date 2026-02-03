package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"jseer/internal/config"
	"jseer/internal/game"
	"jseer/internal/gateway"
	"jseer/internal/logging"
	"jseer/internal/ops"
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

	gw := gateway.New(cfg.Gateway, logger)
	ops.StartAdminServer(cfg.Gateway.AdminAddress, cfg.Gateway.AdminPprof, logger)
	game.RegisterHandlers(gw, &game.Deps{
		Logger:   logger,
		GameIP:   cfg.Game.PublicIP,
		GamePort: cfg.Game.Port,
		Store:    store,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := gw.Start(ctx); err != nil {
		logger.Error("gateway stopped", zap.Error(err))
		os.Exit(1)
	}
}
