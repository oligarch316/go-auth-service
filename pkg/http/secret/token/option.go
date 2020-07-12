package token

import (
	"time"

	"github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-skeleton/pkg/config/types"
	"go.uber.org/zap/zapcore"
)

// CacheOption TODO.
type CacheOption func(*Cache)

// WithClient TODO.
func WithClient(client *secret.Client) CacheOption {
	return func(c *Cache) { c.client.Client = client }
}

// WithClientConfig TODO.
func WithClientConfig(cfg secret.ConfigClient) CacheOption {
	return func(c *Cache) { c.client.ConfigClient = cfg }
}

// WithDefaultTTL TODO.
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) { c.ttl = ttl }
}

// WithMaxSize TODO.
func WithMaxSize(maxSize int64) CacheOption {
	return func(c *Cache) { c.cache.MaxSize(maxSize) }
}

// WithOnHitHook TODO.
func WithOnHitHook(hook func()) CacheOption {
	return func(c *Cache) { c.onHit = hook }
}

// WithOnMissHook TODO.
func WithOnMissHook(hook func()) CacheOption {
	return func(c *Cache) { c.onMiss = hook }
}

// ConfigCache TODO.
type ConfigCache struct {
	Client     secret.ConfigClient `json:"client"`
	DefaultTTL ctype.Duration      `json:"defaultTTL"`
	MaxSize    int64               `json:"maxSize"`
}

// DefaultCacheConfig TODO.
func DefaultCacheConfig() ConfigCache {
	return ConfigCache{
		Client:     secret.DefaultClientConfig(),
		DefaultTTL: ctype.Duration{Duration: 24 * time.Hour},
		MaxSize:    10,
	}
}

// Options TODO.
func (cc ConfigCache) Options() []CacheOption {
	return []CacheOption{
		WithClientConfig(cc.Client),
		WithDefaultTTL(cc.DefaultTTL.Duration),
		WithMaxSize(cc.MaxSize),
	}
}

// MarshalLogObject TODO.
func (cc ConfigCache) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddDuration("defaultTTL", cc.DefaultTTL.Duration)
	enc.AddInt64("maxSize", cc.MaxSize)
	return enc.AddObject("client", cc.Client)
}
