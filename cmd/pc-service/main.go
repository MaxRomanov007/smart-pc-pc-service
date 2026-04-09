package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"smart-pc-pc-service/internal/config"
	httpServer "smart-pc-pc-service/internal/http-server"
	"smart-pc-pc-service/internal/lib/logger"
	"smart-pc-pc-service/internal/lib/logger/sl"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustSetupLogger(cfg.Env)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Debug("debug messages enabled")

	srv := httpServer.New(log, cfg.HTTPServer)
	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Error("http server error", sl.Err(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	<-srv.Done()
}
