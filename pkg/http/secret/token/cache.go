package token

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/karlseguin/ccache/v2"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-auth-service/pkg/secret/token"
)

const (
	compactSep = '.'
	defaultTTL = 24 * time.Hour
)

type cache struct {
	*ccache.Configuration
	*ccache.Cache
}

func (c *cache) init() { c.Cache = ccache.New(c.Configuration) }

type client struct {
	secret.ConfigClient
	*secret.Client
}

func (c *client) init() (err error) {
	if c.Client == nil {
		c.Client, err = secret.NewClient(c.ConfigClient)
	}
	return
}

// Cache TODO.
type Cache struct {
	cache  cache
	client client

	ttl           time.Duration
	onHit, onMiss func()
}

// NewCache TODO.
func NewCache(opts ...CacheOption) (*Cache, error) {
	noop := func() {}

	res := &Cache{
		cache:  cache{Configuration: ccache.Configure()},
		client: client{ConfigClient: secret.DefaultClientConfig()},
		ttl:    defaultTTL,
		onHit:  noop,
		onMiss: noop,
	}

	for _, opt := range opts {
		opt(res)
	}

	res.cache.init()
	return res, res.client.init()
}

func (c *Cache) addKey(key jwk.Key) (*token.Validater, error) {
	keyID := key.KeyID()
	if keyID == "" {
		return nil, errors.New("key is missing key id")
	}

	validater, err := token.NewValidater(key)
	if err != nil {
		return nil, err
	}

	c.cache.Set(keyID, validater, c.ttl)
	return validater, nil
}

func (c *Cache) lookupKey(keyID string) (*token.Validater, bool) {
	if item := c.cache.Get(keyID); item != nil && !item.Expired() {
		return item.Value().(*token.Validater), true
	}

	return nil, false
}

func (c *Cache) fetchKey(keyID string) (jwk.Key, error) {
	key, err := c.client.Key(context.TODO(), keyID)
	if err != nil {
		return nil, err
	}

	if key.KeyID() != keyID {
		return nil, errors.New("received mismatched key id")
	}

	return key, nil
}

// Close TODO.
func (c *Cache) Close() error {
	c.cache.Stop()
	return nil
}

// ItemCount TODO.
func (c *Cache) ItemCount() int { return c.cache.ItemCount() }

// Validate TODO.
func (c *Cache) Validate(token string, claims interface{}) error {
	// Extract key id
	keyID, err := extractKID(token)
	if err != nil {
		return err
	}

	// Attempt cache lookup
	validater, found := c.lookupKey(keyID)

	if found {
		// Found in cache => continue
		c.onHit()
	} else {
		// Not found in cache => acquire via client
		c.onMiss()

		key, err := c.fetchKey(keyID)
		if err != nil {
			return err
		}

		if validater, err = c.addKey(key); err != nil {
			return err
		}
	}

	// Perform validation
	return validater.Validate(token, claims)
}

// Warm TODO.
func (c *Cache) Warm(ctx context.Context) error {
	set, err := c.client.Set(ctx)
	if err != nil {
		return err
	}

	for _, key := range set.Keys {
		if _, err := c.addKey(key); err != nil {
			return err
		}
	}

	return nil
}

func extractKID(token string) (string, error) {
	/*
	   NOTE: Extract from RFC 7519 (https://tools.ietf.org/html/rfc7519)

	   >
	   | JWTs are always represented using the JWS Compact Serialization or the
	   | JWE Compact Serialization.
	*/

	idx := bytes.IndexRune([]byte(token), compactSep)
	if idx < 0 {
		return "", errors.New("invalid token separater format")
	}

	header, err := base64.RawURLEncoding.DecodeString(token[:idx])
	if err != nil {
		return "", errors.New("failed to base64 decode token headers")
	}

	var data struct {
		KeyID string `json:"kid"`
	}
	if err := json.Unmarshal(header, &data); err != nil {
		return "", fmt.Errorf("failed to json decode token headers: %w", err)
	}

	if data.KeyID == "" {
		return "", errors.New("token missing key id (kid)")
	}

	return data.KeyID, nil
}
