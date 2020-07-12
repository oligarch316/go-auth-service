package secret

// import (
//     "time"
//
//     "github.com/oligarch316/go-skeleton/pkg/config/types"
//     "github.com/oligarch316/go-auth-service/pkg/http/secret"
// )
//
// // Option TODO.
// type Option func(*Cache)
//
// // WithClient TODO.
// func WithClient(client *secret.Client) Option {
//     return func(c *Cache) { c.client.Client = client }
// }
//
// // WithClientConfig TODO.
// func WithClientConfig(cfg secret.ConfigClient) Option {
//     return func(c *Cache) { c.client.ConfigClient = cfg }
// }
//
// // WithDefaultTTL TODO.
// func WithDefaultTTL(ttl time.Duration) Option {
//     return func(c *Cache) { c.ttl = ttl }
// }
//
// // WithMaxSize TODO.
// func WithMaxSize(maxSize int64) Option {
//     return func(c *Cache) { c.cache.MaxSize(maxSize) }
// }
//
// // WithOnHitHook TODO.
// func WithOnHitHook(hook func()) Option {
//     return func(c *Cache) { c.onHit = hook }
// }
//
// // WithOnMissHook TODO.
// func WithOnMissHook(hook func()) Option {
//     return func(c *Cache) { c.onMiss = hook }
// }
//
// // Config TODO.
// type Config struct {
//     Client secret.ConfigClient `json:"client"`
//     DefaultTTL ctype.Duration `json:"defaultTTL"`
//     MaxSize int64 `json:"maxSize"`
// }
//
// // DefaultConfig TODO.
// func DefaultConfig() Config {
//     return Config{
//         Client: secret.DefaultClientConfig(),
//         DefaultTTL: ctype.Duration{ Duration: 24 * time.Hour },
//         MaxSize: 10,
//     }
// }
//
// // Options TODO.
// func (c Config) Options() []Option {
//     return []Option{
//         WithClientConfig(c.Client),
//         WithDefaultTTL(c.DefaultTTL.Duration),
//         WithMaxSize(c.MaxSize),
//     }
// }
