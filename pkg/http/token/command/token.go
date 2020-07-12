package command

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/oligarch316/go-auth-service/pkg/http"
	httptoken "github.com/oligarch316/go-auth-service/pkg/http/token"
	"github.com/oligarch316/go-auth-service/pkg/secret/token"
	"github.com/oligarch316/go-auth-service/pkg/store"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// New TODO.
func New(name string, ns *namespace.NS) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Run token http server",
		Long:  "Run token http server",
		Run:   func(_ *cobra.Command, _ []string) { os.Exit(run(ns)) },
	}
}

// NewServer TODO.
func NewServer(cfg Config, srvlet *httpsvc.Servelet, db store.Backend) (*httptoken.Server, error) {
	secretKey, err := cfg.Secret.PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load secret key: %w", err)
	}

	signer, err := token.NewSigner(secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token signer: %w", err)
	}

	return &httptoken.Server{
		ConfigServer: cfg.ConfigServer,
		Servelet:     srvlet,
		Secret:       signer,
		Store:        db,
	}, nil
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

	// ----- HTTP Server
	var (
		servelet = &httpsvc.Servelet{Corelet: observCore.Named("server")}
		router   = httpsvc.NewRouter(servelet, "/")
	)

	server, err := NewServer(cfg.TokenSvc, servelet.Named("token"), db)
	if err != nil {
		root.Logger.Error("failed to create server", zap.Error(err))
		return 1
	}

	server.AddRoutes(router)
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
