package command

import (
	httpsecret "github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-auth-service/pkg/secret"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
)

// Config TODO.
type Config struct {
	Keys []secret.Config `json:"keys"`
}

// DefaultConfig TODO.
func DefaultConfig() Config { return Config{} }

type cmdConfig struct {
	Address   string        `json:"address"`
	Observ    observ.Config `json:"observ"`
	SecretSvc Config        `json:"secretsvc"`
}

func defaultCmdConfig() cmdConfig {
	return cmdConfig{
		Address:   httpsecret.DefaultAddress,
		Observ:    observ.DefaultConfig(),
		SecretSvc: DefaultConfig(),
	}
}

// NewConfig TODO.
func NewConfig(name string, ns *namespace.NS) *cobra.Command {
	cfg := defaultCmdConfig()
	return ns.NewCommand(name, &cfg)
}
