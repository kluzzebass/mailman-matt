package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/exp/slog"
)

type ScheduleFetcher struct {
	cfg    Config
	client http.Client
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
	rawSchedule, err := f.fetchSchedule(ctx, postCode)
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

func (f *ScheduleFetcher) fetchSchedule(ctx context.Context, postCode int) (*rawSchedule, error) {
	// create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.cfg.ApiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	// add headers
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

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
