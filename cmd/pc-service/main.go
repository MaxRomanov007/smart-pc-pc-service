package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"smart-pc-pc-service/internal/config"
	httpServer "smart-pc-pc-service/internal/http-server"
	"smart-pc-pc-service/internal/lib/logger"
	"smart-pc-pc-service/internal/mqtt"
	"smart-pc-pc-service/internal/storage/postgres"

	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/MaxRomanov007/smart-pc-go-lib/waitable"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustSetupLogger(cfg.Env)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Debug("debug messages enabled")

	storage, err := postgres.New(ctx, log, *cfg)
	if err != nil {
		log.Error("failed to create postgres storage", sl.Err(err))
		os.Exit(1)
	}

	srv := httpServer.New(log, cfg.HTTPServer, storage)
	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Error("http server error", sl.Err(err))
			os.Exit(1)
		}
	}()

	conn, err := mqtt.New(ctx, log, cfg.MQTT, storage)
	if err != nil {
		log.Error("failed to create mqtt connection", sl.Err(err))
		os.Exit(1)
	}

	waitable.WaitAll(storage, srv, conn)
}
