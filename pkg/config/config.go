package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/oauth2"
)

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

// APIConfig represents configuration for an instance of Observatorium.
type APIConfig struct {
	URL      string                      `json:"url"`
	Contexts map[TenantName]TenantConfig `json:"contexts"`
}

// TenantConfig represents configuration for a tenant.
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

// Read loads configuration from disk.
func Read(logger log.Logger, path ...string) (*Config, error) {
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

	level.Debug(logger).Log("msg", "read and parsed config file")

	return &cfg, nil
}

// Save writes current config to the disk.
func (c *Config) Save(logger log.Logger) error {
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

	level.Debug(logger).Log("msg", "saved config in config file")

	return nil
}

// AddAPI adds a new Observatorium API to the configuration and saves the config to disk.
// In case no name is provided, the hostname of the API URL is used instead.
func (c *Config) AddAPI(logger log.Logger, name APIName, apiURL string) error {
	if c.APIs == nil {
		c.APIs = make(map[APIName]APIConfig)
		level.Debug(logger).Log("msg", "initialize config API map")
	}

	url, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL", url)
	}

	// url.Parse might pass a URL with only path, so need to check here for scheme and host.
	// As per docs: https://pkg.go.dev/net/url#Parse.
	if url.Host == "" || url.Scheme == "" {
		return fmt.Errorf("%s is not a valid URL", url)
	}

	if name == "" {
		name = APIName(url.Host)
		level.Debug(logger).Log("msg", "use hostname as name")
	}

	if _, ok := c.APIs[name]; ok {
		return fmt.Errorf("api with name %s already exists", name)
	}

	// Add trailing slash if not present.
	parsedUrl := url.String()
	if parsedUrl[len(parsedUrl)-1:] != "/" {
		parsedUrl += "/"
	}

	c.APIs[name] = APIConfig{URL: parsedUrl}

	return c.Save(logger)
}

// RemoveAPI removes a locally saved Observatorium API config as well as its tenants.
// If the current context is pointing to the API being removed, the context is emptied.
func (c *Config) RemoveAPI(logger log.Logger, name APIName) error {
	if _, ok := c.APIs[name]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", name)
	}

	if c.Current.API == name {
		c.Current.API = ""
		c.Current.Tenant = ""
		level.Debug(logger).Log("msg", "empty current config")
	}

	delete(c.APIs, name)

	return c.Save(logger)
}

// AddTenant adds configuration for a tenant under an API and saves it to disk.
// Also, sets new tenant to current in case current config is empty.
func (c *Config) AddTenant(logger log.Logger, name TenantName, api APIName, tenant string, oidcCfg *OIDCConfig) error {
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
		level.Debug(logger).Log("msg", "set new tenant as current")
	}

	return c.Save(logger)
}

// RemoveTenant removes configuration of a tenant under an API and saves changes to disk.
func (c *Config) RemoveTenant(logger log.Logger, name TenantName, api APIName) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[name]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", name, api)
	}

	delete(c.APIs[api].Contexts, name)

	return c.Save(logger)
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
func (c *Config) SetCurrent(logger log.Logger, api APIName, tenant TenantName) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[tenant]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", tenant, api)
	}

	c.Current.API = api
	c.Current.Tenant = tenant

	return c.Save(logger)
}