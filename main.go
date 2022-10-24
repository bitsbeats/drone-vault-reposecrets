package main

import (
	"net/http"

	"github.com/drone/drone-go/plugin/secret"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"github.com/bitsbeats/drone-vault-reposecrets/internal/config"
	"github.com/bitsbeats/drone-vault-reposecrets/pkg/plugin"
)

func main() {
	logger := log.StandardLogger()

	cfg := &config.Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		logger.WithError(err).Fatalf("unable to parse environment")
	}

	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
		logger.Debugf("enabled debug log")
	}

	p, err := plugin.New(
		plugin.WithLogger(logger),
		plugin.WithSecretPath(cfg.VaultSecretPath),
		plugin.WithVaultAddr(cfg.VaultAddr),
		plugin.WithVaultRoleID(cfg.VaultRoleID),
		plugin.WithVaultSecretID(cfg.VaultSecretID),
	)
	if err != nil {
		logger.WithError(err).Fatalf("unable to load plugin")
	}
	handler := secret.Handler(cfg.Secret, p, logger)

	http.Handle("/", handler)
	log.WithField("addr", cfg.Listen).Infof("listening")
	err = http.ListenAndServe(cfg.Listen, nil)
	if err != nil {
		logger.WithError(err).Fatalf("unable to listen")
	}
}
