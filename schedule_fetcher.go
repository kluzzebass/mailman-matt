package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/net/html"
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
	NextDeliveryDays   []string `json:"nextDeliveryDays"`
	IsStreetAddressReq bool     `json:"isStreetAddressReq"`
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
	apiUrl, err := f.fetchAPIURL(ctx)
	if err != nil {
		return nil, err
	}

	rawSchedule, err := f.fetchSchedule(ctx, *apiUrl, postCode)
	if err != nil {
		return nil, err
	}

	schedule, err := parseRawSchedule(ctx, rawSchedule)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (f ScheduleFetcher) fetchAPIURL(ctx context.Context) (*url.URL, error) {
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

func (f *ScheduleFetcher) fetchSchedule(ctx context.Context, apiUrl url.URL, postCode int) (*rawSchedule, error) {

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

	return &raw, nil
}

func parseRawSchedule(ctx context.Context, raw *rawSchedule) (schedule, error) {
	sch := schedule{}

	// parse each date
	for _, date := range raw.NextDeliveryDays {
		t, err := parseDate(date)
		if err != nil {
			return nil, err
		}
		sch = append(sch, *t)
	}

	return sch, nil
}

func parseDate(date string) (*time.Time, error) {

	// parse string using regex and return array of matches
	re := regexp.MustCompile(`(\d+)\.\s+(\w+)$`)
	matches := re.FindStringSubmatch(date)
	if matches == nil || len(matches) != 3 {
		return nil, fmt.Errorf("could not parse date: %s", date)
	}

	// get the month number from the month name
	month, ok := months[matches[2]]
	if !ok {
		return nil, fmt.Errorf("could not parse month: %s", matches[2])
	}

	// parse the day
	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	// parsing dates relative to the current time
	now := time.Now()

	// truncate localized time to midnight so that we can compare dates
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// create a time object with the parsed date based on the current time, since
	// the API only returns the day and month
	parsed := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())

	// if the parsed time is before now, add a year
	if parsed.Before(now) {
		parsed = parsed.AddDate(1, 0, 0)
	}

	return &parsed, nil
}
