package main

import (
	"context"

	// AutoLoad .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	cfg := NewConfig(ctx)
	if cfg.Development {
		cfg.Dump()
	}

}

// func fetchSchedule(ctx)
