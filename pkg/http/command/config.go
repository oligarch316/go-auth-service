package command

import (
	secret "github.com/oligarch316/go-auth-service/pkg/http/secret/command"
	token "github.com/oligarch316/go-auth-service/pkg/http/token/command"
	user "github.com/oligarch316/go-auth-service/pkg/http/user/command"
	"github.com/oligarch316/go-auth-service/pkg/store"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
)

const defaultAddress = "localhost:8000"

type cmdConfig struct {
	Address   string        `json:"address"`
	DB        store.Config  `json:"db"`
	Observ    observ.Config `json:"observ"`
	SecretSvc secret.Config `json:"secretsvc"`
	TokenSvc  token.Config  `json:"tokensvc"`
	UserSvc   user.Config   `json:"usersvc"`
}

func defaultCmdConfig() cmdConfig {
	userSvcConfig := user.DefaultConfig()
	userSvcConfig.SecretCache.Client.Address = defaultAddress

	return cmdConfig{
		Address:   defaultAddress,
		DB:        store.DefaultConfig(),
		Observ:    observ.DefaultConfig(),
		SecretSvc: secret.DefaultConfig(),
		TokenSvc:  token.DefaultConfig(),
		UserSvc:   userSvcConfig,
	}
}

// NewConfig TODO.
func NewConfig(name string, ns *namespace.NS) *cobra.Command {
	cfg := defaultCmdConfig()
	return ns.NewCommand(name, &cfg)
}
