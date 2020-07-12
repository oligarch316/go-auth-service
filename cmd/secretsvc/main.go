package main

import (
	"github.com/oligarch316/go-auth-service/pkg/http/secret/command"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
)

var ns = namespace.New("secretsvc")

func main() {
	var (
		secretCmd  = command.New(ns.Name, ns)
		configCmd  = command.NewConfig("config", ns)
		versionCmd = command.NewVersion("version")
	)

	ns.SetFlags(secretCmd.Flags())
	secretCmd.AddCommand(configCmd, versionCmd)
	secretCmd.Execute()
}
