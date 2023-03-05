package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"
)

type (
	Config struct {
		DumpConfig     bool          `env:"DEVELOPMENT,default=false"`
		Port           int           `env:"MATT_PORT,default=3000"`
		PageUrl        url.URL       `env:"MATT_PAGE_URL,default=https://www.posten.no/levering-av-post"`
		ApiElement     string        `env:"MATT_API_ELEMENT,default=div"`
		ApiElementId   string        `env:"MATT_API_ELEMENT_ID,default=mailbox-delivery-date"`
		ApiPathAttr    string        `env:"MATT_API_URL_ATTR,default=data-component-url"`
		ApiTimeout     time.Duration `env:"MATT_API_TIMEOUT,default=5s"`
		ApiPostCodeArg string        `env:"MATT_API_POSTCODE_ARG,default=postCode"`
		Domain         string        `env:"MATT_DOMAIN,default=example.com"`
		Company        string        `env:"MATT_COMPANY,default=Acme Inc."`
		Product        string        `env:"MATT_COMPANY,default=Example Product"`
		Summary        string        `env:"MATT_SUMMARY,default=POST"`
		Timezone       string        `env:"MATT_TIMEZONE,default=Europe/Oslo"`
		Name           string        `env:"MATT_NAME,default=Matt"`
		CacheTTL       time.Duration `env:"MATT_CACHE_TTL,default=600s"`
	}
)

func NewConfig(ctx context.Context) Config {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}
	if cfg.DumpConfig {
		cfg.Dump()
	}
	return cfg
}

func (c Config) Dump() {
	j, _ := json.MarshalIndent(c, "", "  ")
	log.Println(string(j))
}
