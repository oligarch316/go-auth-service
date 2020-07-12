package command

import (
	"github.com/oligarch316/go-auth-service/internal/pkg/claims"
	"github.com/oligarch316/go-auth-service/pkg/http/secret/token"
	httptoken "github.com/oligarch316/go-auth-service/pkg/http/token"
	"github.com/oligarch316/go-auth-service/pkg/http/user"
	"github.com/oligarch316/go-auth-service/pkg/store"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/config/types"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
)

// Config TODO.
type Config struct {
	AllowedIssuers ctype.StringSet               `json:"allowedIssuers"`
	AudienceNames  httptoken.ConfigAudienceNames `json:"audienceNames"`
	SecretCache    token.ConfigCache             `json:"secretCache"`
}

// DefaultConfig TODO.
func DefaultConfig() Config {
	return Config{
		AllowedIssuers: ctype.NewStringSet(claims.DefaultIssuerName),
		AudienceNames:  httptoken.DefaultAudienceNamesConfig(),
		SecretCache:    token.DefaultCacheConfig(),
	}
}

type cmdConfig struct {
	Address string        `json:"address"`
	DB      store.Config  `json:"db"`
	Observ  observ.Config `json:"observ"`
	UserSvc Config        `json:"usersvc"`
}

func defaultCmdConfig() cmdConfig {
	return cmdConfig{
		Address: user.DefaultAddress,
		DB:      store.DefaultConfig(),
		Observ:  observ.DefaultConfig(),
		UserSvc: DefaultConfig(),
	}
}

// NewConfig TODO.
func NewConfig(name string, ns *namespace.NS) *cobra.Command {
	cfg := defaultCmdConfig()
	return ns.NewCommand(name, &cfg)
}
