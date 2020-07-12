package command

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/oligarch316/go-auth-service/pkg/http"
	httpsecret "github.com/oligarch316/go-auth-service/pkg/http/secret"
	"github.com/oligarch316/go-auth-service/pkg/secret"
	"github.com/oligarch316/go-skeleton/pkg/config/namespace"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// New TODO.
func New(name string, ns *namespace.NS) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Run secret http server",
		Long:  "Run secret http server",
		Run:   func(_ *cobra.Command, _ []string) { os.Exit(run(ns)) },
	}
}

// NewServer TODO.
func NewServer(cfg Config, srvlet *httpsvc.Servelet) (*httpsecret.Server, error) {
	var pubKeys []jwk.Key

	for _, item := range cfg.Keys {
		pubKey, err := item.PublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}

		// Double check
		if !secret.IsPublic(pubKey) {
			return nil, fmt.Errorf("non-public key detected of type %s", pubKey.KeyType().String())
		}

		pubKeys = append(pubKeys, pubKey)
	}

	if len(pubKeys) < 1 {
		return nil, fmt.Errorf("no configured public keys found")
	}

	return &httpsecret.Server{
		Servelet: srvlet,
		Set:      &jwk.Set{Keys: pubKeys},
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

	// ----- HTTP Server
	var (
		servelet = &httpsvc.Servelet{Corelet: observCore.Named("server")}
		router   = httpsvc.NewRouter(servelet, "/")
	)

	server, err := NewServer(cfg.SecretSvc, servelet.Named("secret"))
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
