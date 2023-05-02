package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"golang.org/x/exp/slog"
)

type ScheduleFetcher struct {
	cfg    Config
	client http.Client
}

var months = map[string]int{
	"januar":    1,
	"februar":   2,
	"mars":      3,
	"april":     4,
	"mai":       5,
	"juni":      6,
	"juli":      7,
	"august":    8,
	"september": 9,
	"oktober":   10,
	"november":  11,
	"desember":  12,
}

type rawSchedule struct {
	DeliveryDates []string `json:"delivery_dates"`
}

type schedule []time.Time

func NewScheduleFetcher(cfg Config) *ScheduleFetcher {
	f := &ScheduleFetcher{
		cfg:    cfg,
		client: http.Client{},
	}

	return f
}

func (f *ScheduleFetcher) GetSchedule(ctx context.Context, postCode int) (schedule, error) {
	apiUrl, apiKey, err := f.fetchAPIURL(ctx)
	if err != nil {
		return nil, err
	}

	rawSchedule, err := f.fetchSchedule(ctx, *apiUrl, *apiKey, postCode)
	if err != nil {
		return nil, err
	}

	schedule := schedule{}

	for _, date := range rawSchedule.DeliveryDates {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, t)
	}

	slog.Debug("parsed schedule", "schedule", schedule)

	return schedule, nil
}

func (f ScheduleFetcher) fetchAPIURL(ctx context.Context) (*url.URL, *string, error) {
	res, err := http.Get(f.cfg.PageUrl.String())

	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	// there's no point doing this properly using a token parser, as the page changes all the time

	// regex match for this string:
	// "serviceUrl":"https://www.posten.no/levering-av-post/_/service/no.posten.website/delivery-days"
	re1 := regexp.MustCompile(`"serviceUrl":"([^"]+)"`)
	matches := re1.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return nil, nil, fmt.Errorf("could not find serviceUrl in page")
	}

	apiUrl, err := url.Parse(matches[1])
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("found apiUrl", "api_url", apiUrl.String())

	// regex match for this string:
	// "apiKey":"e3640b22MTY4MzAzNTY5Mg"
	re2 := regexp.MustCompile(`"apiKey":"([^"]+)"`)
	matches = re2.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return nil, nil, fmt.Errorf("could not find apiKey in page")
	}

	apiKey := matches[1]
	slog.Debug("found apiKey", "api_key", apiKey)

	return apiUrl, &apiKey, nil
}

func (f *ScheduleFetcher) fetchSchedule(ctx context.Context, apiUrl url.URL, apiKey string, postCode int) (*rawSchedule, error) {

	// create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	// add headers
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("kp-api-token", apiKey)

	// add query parameters
	v := make(url.Values)
	v.Add("postalCode", fmt.Sprint(postCode))
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

	fmt.Println(string(body))

	var raw rawSchedule
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	slog.Debug("raw schedule", "raw", raw)

	return &raw, nil
}
