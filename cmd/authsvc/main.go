package main

import (
	"github.com/oligarch316/go-auth-service/pkg/http/command"
	secret "github.com/oligarch316/go-auth-service/pkg/http/secret/command"
	token "github.com/oligarch316/go-auth-service/pkg/http/token/command"
	user "github.com/oligarch316/go-auth-service/pkg/http/user/command"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
)

var ns = namespace.New("authsvc")

func main() {
	var (
		authCmd   = command.New(ns.Name, ns)
		secretCmd = secret.New("secret", ns)
		tokenCmd  = token.New("token", ns)
		userCmd   = user.New("user", ns)
	)

	ns.SetFlags(authCmd.Flags())
	ns.SetFlags(secretCmd.Flags())
	ns.SetFlags(tokenCmd.Flags())
	ns.SetFlags(userCmd.Flags())

	secretCmd.AddCommand(secret.NewConfig("config", ns), secret.NewVersion("version"))
	tokenCmd.AddCommand(token.NewConfig("config", ns), token.NewVersion("version"))
	userCmd.AddCommand(user.NewConfig("config", ns), user.NewVersion("version"))

	authCmd.AddCommand(
		command.NewConfig("config", ns),
		command.NewVersion("version"),
		secretCmd,
		tokenCmd,
		userCmd,
	)

	authCmd.Execute()
}
