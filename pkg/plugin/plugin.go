package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/logger"
	"github.com/drone/drone-go/plugin/secret"
	"github.com/hashicorp/vault/api"
)

type (
	// Plugin is a drone plugin
	Plugin struct {
		log logger.Logger

		secretPath string

		vaultAddr     string
		vaultRoleID   string
		vaultSecretID string
	}

	Option func(*Plugin)

	tokenLoader struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
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

func New(options ...Option) (*Plugin, error) {
	p := &Plugin{}
	for _, option := range options {
		option(p)
	}
	if p.log == nil {
		return nil, fmt.Errorf("no logger specified")
	}
	if p.secretPath == "" {
		return nil, fmt.Errorf("no secret path specified")
	}
	if p.vaultAddr == "" {
		return nil, fmt.Errorf("no vault address specified")
	}
	if p.vaultRoleID == "" {
		return nil, fmt.Errorf("no role_id specified")
	}
	if p.vaultSecretID == "" {
		return nil, fmt.Errorf("no secret_id specified")
	}
	return p, nil
}

// Find is Drones Secret handler
func (p *Plugin) Find(ctx context.Context, secReq *secret.Request) (s *drone.Secret, err error) {
	name := secReq.Name

	var data string
	secret, err := p.loadSecret(ctx, name, secReq.Repo.Slug)
	if err != nil {
		return nil, fmt.Errorf("unable to load secret from vault: %s", err)
	}
	switch casted := secret.(type) {
	case string:
		data = casted
	default:
		bytes, err := json.Marshal(secret)
		if err != nil {
			return nil, fmt.Errorf("unable to json-encode secret: %s", err)
		}
		data = string(bytes)
	}

	return &drone.Secret{
		Name: name,
		Data: data,
		Pull: false,
		Fork: false,
	}, nil
}

// loadSecret loads a secret from Vault
func (p *Plugin) loadSecret(ctx context.Context, name, repo string) (secret interface{}, err error) {
	apiClient, err := api.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create vault client: %s", err)
	}

	token, err := p.loadVaultToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch token: %s", err)
	}
	apiClient.SetToken(token)

	path := fmt.Sprintf(p.secretPath, repo)
	vaultSecret, err := apiClient.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch secret: %s", err)
	}
	if vaultSecret == nil {
		return nil, fmt.Errorf("secret value is nil")
	}

	data, ok := vaultSecret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("only kv2 is supported")
	}

	return data[name], nil
}

// loadVaultToken fetches a VaultToken from Vault using the AppRole
func (p *Plugin) loadVaultToken(ctx context.Context) (token string, err error) {
	client := &http.Client{}
	body := strings.NewReader(fmt.Sprintf(
		`{"role_id": "%s", "secret_id": "%s"}`,
		p.vaultRoleID, p.vaultSecretID,
	))
	url := fmt.Sprintf("%s/v1/auth/approle/login", p.vaultAddr)
	tokenReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("unable to create request to fetch vault token: %s", err)
	}
	tokenReq.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("unable to fetch vault token: %s", err)
	}

	tl := &tokenLoader{}
	err = json.NewDecoder(resp.Body).Decode(tl)
	if err != nil {
		return "", fmt.Errorf("unable to parse vault token: %s", err)
	}
	return tl.Auth.ClientToken, nil
}
