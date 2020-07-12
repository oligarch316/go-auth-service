package version

import (
	"github.com/oligarch316/go-skeleton/pkg/version"
	"github.com/oligarch316/go-skeleton/pkg/version/command"

	"github.com/spf13/cobra"
)

var (
	appVersion  string
	gitRevision string

	// GInfo stores global version information.
	GInfo version.Info
)

func init() { GInfo.Init(appVersion, gitRevision) }

// APIInfo TODO
type APIInfo struct {
	version.Info
	APIVersion string
}

// NewCommand TODO
func (ai APIInfo) NewCommand(name string) *cobra.Command {
	return command.New(name, ai, command.WithAppendFields(
		command.Field{
			Name:     "APIVersion",
			Label:    "API Version",
			LongFlag: "api",
		},
	))
}
