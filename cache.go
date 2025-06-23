package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/viccon/sturdyc"
)

type baseConfig struct {
	Capacity           int
	NumShards          int
	TTL                time.Duration
	EvictionPercentage int
	EarlyRefreshes     earlyRefreshConfig
}
type earlyRefreshConfig struct {
	MinRefreshDelay         time.Duration
	MaxRefreshDelay         time.Duration
	SynchronousRefreshDelay time.Duration
	RetryBaseDelay          time.Duration
}

// defaults is the default configuration for the cache.
var defaults = &baseConfig{
	Capacity:           10000,
	NumShards:          10,
	TTL:                10 * time.Second,
	EvictionPercentage: 10,
	EarlyRefreshes: earlyRefreshConfig{
		MinRefreshDelay:         time.Millisecond * 10,
		MaxRefreshDelay:         time.Millisecond * 30,
		SynchronousRefreshDelay: time.Second * 10,
		RetryBaseDelay:          time.Millisecond * 10,
	},
}

// Option is a functional Option for configuring the cache.
type Option func(*baseConfig)

// WithTTL sets the TTL for the cache.
func WithTTL(ttl time.Duration) Option {
	return func(c *baseConfig) {
		c.TTL = ttl
	}
}

// NewCache creates a new cache with the given name and store.
func NewCache(cacheName string, client *redis.Client, opts ...Option) *sturdyc.Client[any] {
	for _, opt := range opts {
		opt(defaults)
	}

	cacheInstance := sturdyc.New[any](
		defaults.Capacity,
		defaults.NumShards,
		defaults.TTL,
		defaults.EvictionPercentage,
		sturdyc.WithEarlyRefreshes(
			defaults.EarlyRefreshes.MinRefreshDelay,
			defaults.EarlyRefreshes.MaxRefreshDelay,
			defaults.EarlyRefreshes.SynchronousRefreshDelay,
			defaults.EarlyRefreshes.RetryBaseDelay,
		),
		sturdyc.WithRefreshCoalescing(3, defaults.TTL),
		sturdyc.WithDistributedStorageEarlyRefreshes(newRedisStore(client, defaults.TTL), 10*time.Second),
	)

	return cacheInstance
}

var _ sturdyc.DistributedStorageWithDeletions = &redisStore{}

type redisStore struct {
	*redis.Client
	ttl time.Duration
}

func newRedisStore(client *redis.Client, ttl time.Duration) *redisStore {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &redisStore{
		Client: client,
		ttl:    ttl,
	}
}

// Get implements sturdyc.DistributedStorage.
func (r *redisStore) Get(ctx context.Context, key string) ([]byte, bool) {
	val, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		slog.DebugContext(ctx, "failed to get key from redis", "key", key, "err", err)
		return nil, false
	}
	return val, true
}

// GetBatch implements sturdyc.DistributedStorage.
func (r *redisStore) GetBatch(ctx context.Context, keys []string) map[string][]byte {
	records := make(map[string][]byte)
	for _, key := range keys {
		val, ok := r.Get(ctx, key)
		if ok {
			records[key] = val
		}
	}
	return records
}

// Set implements sturdyc.DistributedStorage.
func (r *redisStore) Set(ctx context.Context, key string, value []byte) {
	// default to 5 minutes lifetime
	if err := r.Client.Set(ctx, key, value, r.ttl).Err(); err != nil {
		slog.DebugContext(ctx, "failed to set key in redis", "key", key, "err", err)
	}
}

// SetBatch implements sturdyc.DistributedStorage.
func (r *redisStore) SetBatch(ctx context.Context, records map[string][]byte) {
	for key, val := range records {
		r.Set(ctx, key, val)
	}
}

// Delete implements sturdyc.DistributedStorageWithDeletions.
func (r *redisStore) Delete(ctx context.Context, key string) {
	if err := r.Client.Del(ctx, key).Err(); err != nil {
		slog.DebugContext(ctx, "failed to delete key in redis", "key", key, "err", err)
	}
}

// DeleteBatch implements sturdyc.DistributedStorageWithDeletions.
func (r *redisStore) DeleteBatch(ctx context.Context, keys []string) {
	if err := r.Client.Del(ctx, keys...); err != nil {
		slog.DebugContext(ctx, "failed to delete keys in redis", "keys", keys, "err", err)
	}
}
