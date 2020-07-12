package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/goware/urlx"
	"github.com/lestrrat-go/jwx/jwk"
	"go.uber.org/zap/zapcore"
)

const defaultURLScheme = "http"

// ConfigClient TODO.
type ConfigClient struct {
	Address string        `json:"address"`
	Timeout time.Duration `json:"timeout"`
}

// DefaultClientConfig TODO.
func DefaultClientConfig() ConfigClient {
	return ConfigClient{
		Address: DefaultAddress,
		Timeout: 10 * time.Second,
	}
}

// MarshalLogObject TODO.
func (cc ConfigClient) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", cc.Address)
	enc.AddDuration("timeout", cc.Timeout)
	return nil
}

// Client TODO.
type Client struct {
	client         *http.Client
	urlSet, urlKey string
}

// NewClient TODO.
func NewClient(cfg ConfigClient) (*Client, error) {
	base, err := urlx.ParseWithDefaultScheme(cfg.Address, defaultURLScheme)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	base.Path = urlPathJoin(APIVersion, pathBase)

	return &Client{
		client: &http.Client{Timeout: cfg.Timeout},
		urlSet: urlPathJoin(base.String(), pathSet),
		urlKey: urlPathJoin(base.String(), pathKey),
	}, nil
}

func (c *Client) do(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	cTypeJSON.writeReqHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// Set TODO.
func (c *Client) Set(ctx context.Context) (*jwk.Set, error) {
	respBytes, err := c.do(ctx, c.urlSet)
	if err != nil {
		return nil, err
	}

	return jwk.ParseBytes(respBytes)
}

// KeyIDs TODO.
func (c *Client) KeyIDs(ctx context.Context) ([]string, error) {
	respBytes, err := c.do(ctx, c.urlKey)
	if err != nil {
		return nil, err
	}

	var list keyListResponse
	if err := json.Unmarshal(respBytes, &list); err != nil {
		return nil, err
	}

	return list.KeyIDs, nil
}

// Key TODO.
func (c *Client) Key(ctx context.Context, id string) (jwk.Key, error) {
	respBytes, err := c.do(ctx, urlPathJoin(c.urlKey, id))
	if err != nil {
		return nil, err
	}

	return jwk.ParseKey(respBytes)
}

// NOTE: path.Join(...) ruins urls by also applying path.Clean(...)
func urlPathJoin(elems ...string) string { return strings.Join(elems, "/") }
