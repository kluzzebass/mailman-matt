package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/jellydator/ttlcache/v3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/slog"
)

type WebServer struct {
	cfg     Config
	fetcher *ScheduleFetcher
	builder *CalendarBuilder
	cache   *ttlcache.Cache[int, *ics.Calendar]
	echo    *echo.Echo
	logger  *slog.Logger
}

func NewWebServer(cfg Config, fetcher *ScheduleFetcher, builder *CalendarBuilder) *WebServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	logger := slog.Default().With("subsystem", "http")

	s := &WebServer{
		cfg:     cfg,
		fetcher: fetcher,
		builder: builder,
		cache:   NewCache(),
		echo:    e,
		logger:  logger,
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogStatus:        true,
		LogLatency:       true,
		LogProtocol:      true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     true,
		LogReferer:       true,
		LogUserAgent:     true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				"uri", v.URI,
				"uri_path", v.URIPath,
				"status", v.Status,
				"method", v.Method,
				"latency", v.Latency,
				"latency_human", v.Latency.String(),
				"remote_ip", v.RemoteIP,
				"host", v.Host,
				"referer", v.Referer,
				"user_agent", v.UserAgent,
				"request_id", v.RequestID,
				"error", v.Error,
				"response_size", v.ResponseSize,
				"route_path", v.RoutePath,
				"content_length", v.ContentLength,
				"protocol", v.Protocol,
				"start_time", v.StartTime,
			)

			return nil
		},
	}))

	// 4 digit postcode regex matcher
	re := regexp.MustCompile(`^\d{4}$`)

	// GET /:postCode
	e.GET("/:postCode", func(c echo.Context) error {
		ctx := c.Request().Context()
		postCode := c.Param("postCode")
		if !re.MatchString(postCode) {
			return c.String(http.StatusBadRequest, "invalid postcode")
		}

		// convert postCode to int
		postCodeInt, _ := strconv.Atoi(postCode)

		var cal *ics.Calendar

		if item := s.cache.Get(postCodeInt); item != nil {
			logger.Info("hit",
				"subsystem", "cache",
				"item", item.Key(),
			)
			cal = item.Value()
		} else {
			sch, err := fetcher.GetSchedule(ctx, postCodeInt)
			if err != nil {
				s.logger.Error("error fetching schedule", err)
				return c.String(http.StatusGatewayTimeout, "operation timed out while fetching schedule")
			}

			// build ics calendar from schedule and serialize it
			cal = builder.buildCalendar(sch)

			// calculate the cache ttl duration until midnight
			now := time.Now()
			tomorrow := now.AddDate(0, 0, 1)
			tomorrowMidnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location())
			ttl := tomorrowMidnight.Sub(now)

			// print duration until midnight
			logger.Debug("ttl", "subsystem", "cache", "postCode", postCode, "duration", ttl.String())

			s.cache.Set(postCodeInt, cal, ttl)
		}

		c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
		c.Response().Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
		c.Response().Header().Set("Date", time.Now().String())
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Keep-Alive", "timeout=5, max=1000")

		return c.String(http.StatusOK, cal.Serialize())
	})

	return s
}

func (s *WebServer) Start() error {
	go func() {
		addr := fmt.Sprintf(":%d", s.cfg.Port)
		s.logger.Info("starting web server", "addr", addr)
		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.Info("shutting down the server")
		}
	}()

	return nil
}

func (s *WebServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping the server")
	return s.echo.Shutdown(ctx)
}
