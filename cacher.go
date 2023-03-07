package main

import (
	"context"

	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/exp/slog"
)

func NewCache() *ttlcache.Cache[int, string] {

	logger := slog.Default().With("subsystem", "cache")
	cache := ttlcache.New[int, string]()

	cache.OnInsertion(func(ctx context.Context, item *ttlcache.Item[int, string]) {
		logger.Debug("insertion",
			"item", item.Key(),
			"expires_at", item.ExpiresAt(),
		)
	})

	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[int, string]) {
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
