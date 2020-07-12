package main

import (
	"github.com/oligarch316/go-auth-service/pkg/http/token/command"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
)

var ns = namespace.New("tokensvc")

func main() {
	var (
		tokenCmd   = command.New(ns.Name, ns)
		configCmd  = command.NewConfig("config", ns)
		versionCmd = command.NewVersion("version")
	)

	ns.SetFlags(tokenCmd.Flags())
	tokenCmd.AddCommand(configCmd, versionCmd)
	tokenCmd.Execute()
}
