package main

import (
	"context"

	ics "github.com/arran4/golang-ical"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/exp/slog"
)

func NewCache() *ttlcache.Cache[int, *ics.Calendar] {

	logger := slog.Default().With("subsystem", "cache")
	cache := ttlcache.New[int, *ics.Calendar]()

	cache.OnInsertion(func(ctx context.Context, item *ttlcache.Item[int, *ics.Calendar]) {
		logger.Debug("insertion",
			"item", item.Key(),
			"expires_at", item.ExpiresAt(),
		)
	})

	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[int, *ics.Calendar]) {
		reasonStr := "unknown"

		switch reason {
		case ttlcache.EvictionReasonCapacityReached:
			reasonStr = "capacity reached"
		case ttlcache.EvictionReasonExpired:
			reasonStr = "expired"
		case ttlcache.EvictionReasonDeleted:
			reasonStr = "deleted"
		}
		logger.Debug("eviction",
			"item", item.Key(),
			"reason", reasonStr,
		)
	})

	// crank up the auto evictor
	go cache.Start()

	return cache
}
