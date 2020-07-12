package command

import (
	"github.com/oligarch316/go-auth-service/pkg/http/token"
	"github.com/oligarch316/go-auth-service/pkg/secret"
	"github.com/oligarch316/go-auth-service/pkg/store"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
)

// Config TODO.
type Config struct {
	token.ConfigServer
	Secret secret.Config `json:"secret"`
}

// DefaultConfig TODO.
func DefaultConfig() Config {
	return Config{
		ConfigServer: token.DefaultServerConfig(),
		Secret:       secret.DefaultConfig(),
	}
}

type cmdConfig struct {
	Address  string        `json:"address"`
	DB       store.Config  `json:"db"`
	Observ   observ.Config `json:"observ"`
	TokenSvc Config        `json:"tokensvc"`
}

func defaultCmdConfig() cmdConfig {
	return cmdConfig{
		Address:  token.DefaultAddress,
		DB:       store.DefaultConfig(),
		Observ:   observ.DefaultConfig(),
		TokenSvc: DefaultConfig(),
	}
}

// NewConfig TODO.
func NewConfig(name string, ns *namespace.NS) *cobra.Command {
	cfg := defaultCmdConfig()
	return ns.NewCommand(name, &cfg)
}
