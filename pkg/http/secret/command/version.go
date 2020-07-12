package command

import (
	"github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-auth-service/pkg/version"
	"github.com/spf13/cobra"
)

var versionInfo = version.APIInfo{
	APIVersion: secret.APIVersion,
	Info:       version.GInfo,
}

// NewVersion TODO.
func NewVersion(name string) *cobra.Command { return versionInfo.NewCommand(name) }
