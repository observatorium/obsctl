package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/oauth2"
)

// TODO: Add tests.

const (
	configFileName = "config.json"
	configDirName  = "obsctl"
)

// getConfigFilePath returns the obsctl config file path or the first string in override argument.
// Useful for testing.
func getConfigFilePath(override ...string) string {
	if len(override) != 0 {
		return override[0]
	}

	usrConfigDir, err := os.UserConfigDir()
	if err != nil {
		return configFileName
	}

	return filepath.Join(usrConfigDir, configDirName, configFileName)
}

func ensureConfigDir() error {
	if err := os.MkdirAll(path.Dir(getConfigFilePath()), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	return nil
}

type APIName string
type TenantName string

// Config represents the structure of the configuration file.
type Config struct {
	pathOverride []string

	APIs    map[APIName]APIConfig `json:"apis"`
	Current struct {
		API    APIName    `json:"api"`
		Tenant TenantName `json:"tenant"`
	} `json:"current"`
}

type APIConfig struct {
	URL      string                      `json:"url"`
	Contexts map[TenantName]TenantConfig `json:"contexts"`
}

type TenantConfig struct {
	Tenant string      `json:"tenant"`
	CAFile []byte      `json:"ca"`
	OIDC   *OIDCConfig `json:"oidc"`
}

// OIDCConfig represents OIDC auth config for a tenant.
type OIDCConfig struct {
	Token *oauth2.Token `json:"token"`

	Audience     string `json:"audience"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	IssuerURL    string `json:"issuerURL"`
}

func Read(path ...string) (*Config, error) {
	if err := ensureConfigDir(); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(getConfigFilePath(path...), os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	cfg := Config{pathOverride: path}

	if err := json.NewDecoder(file).Decode(&cfg); err != nil && err != io.EOF {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	file, err := os.OpenFile(getConfigFilePath(c.pathOverride...), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(c); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// AddAPI adds a new Observatorium API to the configuration and saves the config to disk.
func (c *Config) AddAPI(name APIName, url string) error {
	if c.APIs == nil {
		c.APIs = make(map[APIName]APIConfig)
	}

	if _, ok := c.APIs[name]; ok {
		return fmt.Errorf("api with name %s already exists", name)
	}

	c.APIs[name] = APIConfig{URL: url}

	return c.Save()
}

// RemoveAPI removes a locally saved Observatorium API config as well as its tenants.
// If the current context is pointing to the API being removed, the context is emptied.
func (c *Config) RemoveAPI(name APIName) error {
	if _, ok := c.APIs[name]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", name)
	}

	if c.Current.API == name {
		c.Current.API = ""
		c.Current.Tenant = ""
	}

	delete(c.APIs, name)

	return c.Save()
}

func (c *Config) AddTenant(name TenantName, api APIName, tenant string, oidcCfg *OIDCConfig) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if c.APIs[api].Contexts == nil {
		a := c.APIs[api]
		a.Contexts = make(map[TenantName]TenantConfig)

		c.APIs[api] = a
	}

	if _, ok := c.APIs[api].Contexts[name]; ok {
		return fmt.Errorf("tenant with name %s already exists in api %s", name, api)
	}

	c.APIs[api].Contexts[name] = TenantConfig{
		Tenant: tenant,
		OIDC:   oidcCfg,
	}

	// If the current context is empty, set the newly added tenant as current.
	if c.Current.API == "" && c.Current.Tenant == "" {
		c.Current.API = api
		c.Current.Tenant = name
	}

	return c.Save()
}

func (c *Config) RemoveTenant(name TenantName, api APIName) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[name]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", name, api)
	}

	delete(c.APIs[api].Contexts, name)

	return c.Save()
}

func (c *Config) GetContext(api APIName, tenant TenantName) (TenantConfig, APIConfig, error) {
	if _, ok := c.APIs[api]; !ok {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("api with name %s doesn't exist", c.Current.API)
	}

	if _, ok := c.APIs[api].Contexts[tenant]; !ok {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("tenant with name %s doesn't exist in api %s", c.Current.Tenant, c.Current.API)
	}

	return c.APIs[api].Contexts[tenant], c.APIs[api], nil
}

// GetCurrentContext returns the currently set context i.e, the current API and tenant configuration.
func (c *Config) GetCurrent() (TenantConfig, APIConfig, error) {
	if c.Current.API == "" || c.Current.Tenant == "" {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("current context is empty")
	}

	return c.GetContext(c.Current.API, c.Current.Tenant)
}

// SetCurrent switches the current context to given api and tenant.
func (c *Config) SetCurrent(api APIName, tenant TenantName) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[tenant]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", tenant, api)
	}

	c.Current.API = api
	c.Current.Tenant = tenant

	return c.Save()
}
