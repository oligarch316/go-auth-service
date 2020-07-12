package command

import (
	"github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-auth-service/pkg/http/token"
	"github.com/oligarch316/go-auth-service/pkg/version"
	skelversion "github.com/oligarch316/go-skeleton/pkg/version"
	"github.com/oligarch316/go-skeleton/pkg/version/command"
	"github.com/spf13/cobra"
)

type apiInfo struct {
	skelversion.Info
	SecretAPIVersion, TokenAPIVersion, UserAPIVersion string
}

var versionInfo = apiInfo{
	Info:             version.GInfo,
	SecretAPIVersion: secret.APIVersion,
	TokenAPIVersion:  token.APIVersion,
	UserAPIVersion:   "TODO",
}

// NewVersion TODO.
func NewVersion(name string) *cobra.Command {
	return command.New(name, versionInfo, command.WithAppendFields(
		command.Field{
			Name:     "SecretAPIVersion",
			Label:    "Secret API Version",
			LongFlag: "secret-api",
		},
		command.Field{
			Name:     "TokenAPIVersion",
			Label:    "Token API Version",
			LongFlag: "token-api",
		},
		command.Field{
			Name:     "UserAPIVersion",
			Label:    "User API Version",
			LongFlag: "user-api",
		},
	))
}
