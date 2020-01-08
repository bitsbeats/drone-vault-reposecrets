package plugin

import (
	"github.com/drone/drone-go/plugin/logger"
)

func WithLogger(l logger.Logger) Option {
	return func(p *Plugin) {
		p.log = l
	}
}

func WithSecretPath(path string) Option {
	return func(p *Plugin) {
		p.secretPath = path
	}
}

func WithVaultAddr(addr string) Option {
	return func(p *Plugin) {
		p.vaultAddr = addr
	}
}

func WithVaultRoleID(roleID string) Option {
	return func(p *Plugin) {
		p.vaultRoleID = roleID
	}
}

func WithVaultSecretID(secretID string) Option {
	return func(p *Plugin) {
		p.vaultSecretID = secretID
	}
}
