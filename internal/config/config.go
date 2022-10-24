package config

type (
	// Config holds the plugins configuration
	Config struct {
		Secret string `envconfig:"PLUGIN_SECRET"`
		Listen string `envconfig:"PLUGIN_LISTEN" default:":8080"`
		Debug  bool   `envconfig:"PLUGIN_DEBUG"`

		VaultSecretPath string `envconfig:"VAULT_SECRET_PATH" required:"true"`

		VaultAddr     string `envconfig:"VAULT_ADDR" required:"true"`
		VaultRoleID   string `envconfig:"VAULT_ROLE_ID" required:"true"`
		VaultSecretID string `envconfig:"VAULT_SECRET_ID" required:"true"`
	}
)
