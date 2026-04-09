package main

import (
	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/lib/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustSetupLogger(cfg.Env)

	log.Debug("debug messages enabled")
}
