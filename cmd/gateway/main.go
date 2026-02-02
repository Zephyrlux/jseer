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

	gw := gateway.New(cfg.Gateway, logger)
	game.RegisterHandlers(gw, &game.Deps{Logger: logger})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := gw.Start(ctx); err != nil {
		logger.Error("gateway stopped", zap.Error(err))
		os.Exit(1)
	}
}
