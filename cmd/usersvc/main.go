package main

import (
	"github.com/oligarch316/go-auth-service/pkg/http/user/command"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
)

var ns = namespace.New("usersvc")

func main() {
	var (
		userCmd    = command.New(ns.Name, ns)
		configCmd  = command.NewConfig("config", ns)
		versionCmd = command.NewVersion("version")
	)

	ns.SetFlags(userCmd.Flags())
	userCmd.AddCommand(configCmd, versionCmd)
	userCmd.Execute()
}
