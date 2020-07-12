package command

import (
	"log"
	"net/http"
	"os"

	"github.com/oligarch316/go-auth-service/pkg/http"
	secret "github.com/oligarch316/go-auth-service/pkg/http/secret/command"
	token "github.com/oligarch316/go-auth-service/pkg/http/token/command"
	user "github.com/oligarch316/go-auth-service/pkg/http/user/command"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// New TODO.
func New(name string, ns *namespace.NS) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Run combined secret/token/user http server",
		Long:  "Run combined secret/token/user http server",
		Run:   func(_ *cobra.Command, _ []string) { os.Exit(run(ns)) },
	}
}

func run(ns *namespace.NS) int {
	// ----- Boilerplate
	cfg := defaultCmdConfig()

	srcRecord, err := ns.LoadAndRecord(&cfg)
	if err != nil {
		log.Printf("[bootstrap] failed to load configuration: %s\n", err)
		return 1
	}

	observCore, observCleanup, err := observ.Bootstrap(cfg.Observ)
	if err != nil {
		log.Printf("[bootstrap] failed to build observability core: %s\n", err)
		return 1
	}

	// TODO: error handling
	defer observCleanup()

	root := observCore.Named("root")
	root.Logger.Info("loaded configuration", zap.Strings("sources", srcRecord))
	root.Logger.Info("created observability core", zap.Object("config", cfg.Observ))

	// ----- Database
	db, err := cfg.DB.Build(observCore.Named("database"))
	if err != nil {
		root.Logger.Error("failed to create database", zap.Error(err))
		return 1
	}

	defer db.Close()

	root.Logger.Info("created database", zap.Object("config", cfg.DB))

	// ----- HTTP Servers
	var (
		servelet = &httpsvc.Servelet{Corelet: observCore.Named("server")}
		router   = httpsvc.NewRouter(servelet, "/")
	)

	secretSvr, err := secret.NewServer(cfg.SecretSvc, servelet.Named("secret"))
	if err != nil {
		root.Logger.Error("failed to create secret server", zap.Error(err))
		return 1
	}

	tokenSvr, err := token.NewServer(cfg.TokenSvc, servelet.Named("token"), db)
	if err != nil {
		root.Logger.Error("failed to create token server", zap.Error(err))
		return 1
	}

	userSvr, err := user.NewServer(cfg.UserSvc, servelet.Named("user"), db)
	if err != nil {
		root.Logger.Error("failed to create user server", zap.Error(err))
		return 1
	}

	secretSvr.AddRoutes(router)
	tokenSvr.AddRoutes(router)
	userSvr.AddRoutes(router)
	router.AddMetaRoutes()

	// ----- Run
	root.Logger.Info(
		"starting server",
		zap.String("address", cfg.Address),
		zap.Array("routes", router),
	)

	// TODO: Signal handlers
	// TODO: TLS only

	if err := http.ListenAndServe(cfg.Address, router); err != http.ErrServerClosed {
		root.Logger.Error("server failure", zap.Error(err))
		return 1
	}

	return 0
}
