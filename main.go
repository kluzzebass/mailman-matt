package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	// AutoLoad .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	cfg := NewConfig(ctx)

	level := slog.LevelInfo

	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	ho := slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.LogSource,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &ho))
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
