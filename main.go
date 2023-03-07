package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	// AutoLoad .env file

	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	cfg := NewConfig(ctx)

	ho := slog.HandlerOptions{
		Level:     cfg.LogLevel,
		AddSource: cfg.LogSource,
	}

	logger := slog.New(ho.NewJSONHandler(os.Stdout))
	logger = logger.With("subsystem", "main")

	if cfg.DumpConfig {
		logger.Info("dumping configuration", "config", cfg)
	}

	// look at me... i'm the logger now.
	slog.SetDefault(logger)

	logger.Info("starting service")

	fetcher := NewScheduleFetcher(cfg)
	builder := NewCalendarBuilder(cfg)
	server := NewWebServer(cfg, fetcher, builder)

	// Start server
	server.Start()

	// Wait for interrupt signal to gracefully shutdown the server with
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logger.Warn("error shutting down web server", "error", err)
	}

	logger.Info("ok, bye!")
}
