package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	// AutoLoad .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()
	cfg := NewConfig(ctx)
	_ = cfg

	fetchSchedule(ctx, cfg, 9845)
}

func fetchSchedule(ctx context.Context, cfg Config, postCode int) {
	ctx, _ = context.WithTimeout(ctx, cfg.ApiTimeout)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.ApiUrl.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Set("user-agent", "mailman-matt")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "*")
	req.Header.Set("accept-encoding", "gzip, b")

	v := make(url.Values)
	v.Add("postCode", fmt.Sprint(postCode))
	req.URL.RawQuery = v.Encode()

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		log.Println("Request failed:", err)
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	// Print the status code if the request succeeds
	log.Println("Status code:", res.StatusCode)
	log.Println("Header:", res.Header)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Reading response body failed:", err)
		return
	}

	fmt.Println("Body:", string(body))
}
