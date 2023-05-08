package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"
	"golang.org/x/exp/slog"
)

type (
	Config struct {
		LogLevel   *slog.LevelVar `env:"MATT_LOG_LEVEL,default=info"`
		LogSource  bool           `env:"MATT_LOG_SOURCE,default=false"`
		DumpConfig bool           `env:"MATT_DUMP_CONFIG,default=false"`
		Port       int            `env:"MATT_PORT,default=3000"`
		PageUrl    url.URL        `env:"MATT_PAGE_URL,default=https://www.posten.no/levering-av-post"`
		ApiTimeout time.Duration  `env:"MATT_API_TIMEOUT,default=5s"`
		ProductID  string         `env:"MATT_PRODUCT_ID,default=-//github.com//kluzzebass//mailman-matt-go//EN"`
		Summary    string         `env:"MATT_SUMMARY,default=POST"`
		Timezone   string         `env:"MATT_TIMEZONE,default=Europe/Oslo"`
		Name       string         `env:"MATT_NAME,default=Matt"`
		CacheTTL   time.Duration  `env:"MATT_CACHE_TTL,default=1h"`
	}
)

func NewConfig(ctx context.Context) Config {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

func (c Config) Dump() {
	j, _ := json.MarshalIndent(c, "", "  ")
	log.Println(string(j))
}
