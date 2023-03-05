package main

import (
	"context"
	"log"

	// AutoLoad .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	cfg := NewConfig(ctx)
	fetcher := NewFetcher(cfg)

	// fetcher.fetchSchedule(ctx, 9845)
	err := fetcher.GetSchedule(ctx, 9845)
	if err != nil {
		panic(err)
	}

	log.Println("Done!")
}
