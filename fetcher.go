package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type Fetcher struct {
	cfg    Config
	client http.Client
}

type rawSchedule struct {
	NextDeliveryDays   []string `json:"nextDeliveryDays"`
	IsStreetAddressReq bool     `json:"isStreetAddressReq"`
}

type schedule struct {
	NextDeliveryDays []string
}

func NewFetcher(cfg Config) *Fetcher {
	f := &Fetcher{
		cfg:    cfg,
		client: http.Client{},
	}

	return f
}

func (f *Fetcher) GetSchedule(ctx context.Context, postCode int) error {
	apiUrl, err := f.fetchAPIURL(ctx)
	if err != nil {
		return err
	}

	rawSchedule, err := f.fetchSchedule(ctx, *apiUrl, postCode)
	if err != nil {
		return err
	}

	schedule, err := f.parseRawSchedule(ctx, rawSchedule)
	if err != nil {
		return err
	}

	fmt.Println(schedule)
	return nil
}

func (f Fetcher) fetchAPIURL(ctx context.Context) (*url.URL, error) {
	res, err := http.Get(f.cfg.PageUrl.String())

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	z := html.NewTokenizer(res.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		t := z.Token()

		if t.Type != html.StartTagToken || t.Data != f.cfg.ApiElement {
			continue
		}

		// locate the correct element
		foundElement := false
		for _, a := range t.Attr {
			// fmt.Println(a.Key, a.Val)
			if a.Key == "id" && a.Val == f.cfg.ApiElementId {
				foundElement = true
				break
			}
		}

		if !foundElement {
			continue
		}

		// locate the correct attribute
		for _, a := range t.Attr {
			if a.Key == f.cfg.ApiPathAttr {
				url, err := url.Parse(a.Val)
				if err != nil {
					return nil, err
				}
				// handle both relative and absolute URLs
				url = f.cfg.PageUrl.ResolveReference(url)
				return url, nil
			}
		}
	}
	return nil, io.EOF
}

func (f *Fetcher) fetchSchedule(ctx context.Context, apiUrl url.URL, postCode int) (*rawSchedule, error) {

	// create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	// add headers
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

	// add query parameters
	v := make(url.Values)
	v.Add(f.cfg.ApiPostCodeArg, fmt.Sprint(postCode))
	req.URL.RawQuery = v.Encode()

	// send the request
	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	raw := rawSchedule{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	fmt.Println("Raw Schecule:", raw)

	return &raw, nil
}

func (f *Fetcher) parseRawSchedule(ctx context.Context, raw *rawSchedule) (*schedule, error) {
	return nil, nil
}
