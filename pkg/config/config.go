package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	configFileName = "config.json"
	configDirName  = "obsctl"
	envVar         = "OBSCTL_CONFIG_PATH"
)

// getConfigFilePath returns the obsctl config file path or the value of env variable.
// Useful for testing.
func getConfigFilePath() string {
	override := os.Getenv(envVar)
	if len(override) != 0 {
		return override
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

// Config represents the structure of the configuration file.
type Config struct {
	pathOverride string

	APIs    map[string]APIConfig `json:"apis"`
	Current struct {
		API    string `json:"api"`
		Tenant string `json:"tenant"`
	} `json:"current"`
}

// APIConfig represents configuration for an instance of Observatorium.
type APIConfig struct {
	URL      string                  `json:"url"`
	Contexts map[string]TenantConfig `json:"contexts"`
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

// Client returns a OAuth2 HTTP client based on the configuration for a tenant.
func (t *TenantConfig) Client(ctx context.Context, logger log.Logger) (*http.Client, error) {
	if t.OIDC != nil {
		provider, err := oidc.NewProvider(ctx, t.OIDC.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("constructing oidc provider: %w", err)
		}

		ccc := clientcredentials.Config{
			ClientID:     t.OIDC.ClientID,
			ClientSecret: t.OIDC.ClientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
			Scopes:       []string{"openid", "offline_access"},
		}

		if t.OIDC.Audience != "" {
			ccc.EndpointParams = url.Values{
				"audience": []string{t.OIDC.Audience},
			}
		}

		ts := ccc.TokenSource(ctx)

		// If token has not expired, we can reuse.
		if t.OIDC.Token != nil {
			currentTime := time.Now()
			if t.OIDC.Token.Expiry.After(currentTime) {
				ts = oauth2.ReuseTokenSource(t.OIDC.Token, ts)
			}
		}

		tkn, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("fetching token: %w", err)
		}

		t.OIDC.Token = tkn

		level.Debug(logger).Log("msg", "fetched token", "tenant", t.Tenant)

		return oauth2.NewClient(ctx, ts), nil
	}

	return http.DefaultClient, nil
}

// Tenant returns a OAuth2 HTTP transport based on the configuration for a tenant.
func (t *TenantConfig) Transport(ctx context.Context, logger log.Logger) (http.RoundTripper, error) {
	if t.OIDC != nil {
		provider, err := oidc.NewProvider(ctx, t.OIDC.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("constructing oidc provider: %w", err)
		}

		ccc := clientcredentials.Config{
			ClientID:     t.OIDC.ClientID,
			ClientSecret: t.OIDC.ClientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
			Scopes:       []string{"openid", "offline_access"},
		}

		if t.OIDC.Audience != "" {
			ccc.EndpointParams = url.Values{
				"audience": []string{t.OIDC.Audience},
			}
		}

		ts := ccc.TokenSource(ctx)

		// If token has not expired, we can reuse.
		if t.OIDC.Token != nil {
			currentTime := time.Now()
			if t.OIDC.Token.Expiry.After(currentTime) {
				ts = oauth2.ReuseTokenSource(t.OIDC.Token, ts)
			}
		}

		tkn, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("fetching token: %w", err)
		}

		t.OIDC.Token = tkn

		level.Debug(logger).Log("msg", "fetched token", "tenant", t.Tenant)

		return &oauth2.Transport{
			Source: ts,
			Base:   http.DefaultTransport,
		}, nil
	}

	return http.DefaultTransport, nil
}

// Client returns an OAuth2 HTTP client based on the current context configuration.
func (c *Config) Client(ctx context.Context, logger log.Logger) (*http.Client, error) {
	tenant, _, err := c.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("getting current context: %w", err)
	}

	client, err := tenant.Client(ctx, logger)
	if err != nil {
		return nil, err
	}

	c.APIs[c.Current.API].Contexts[c.Current.Tenant] = tenant
	if err := c.Save(logger); err != nil {
		return nil, fmt.Errorf("updating token in config file: %w", err)
	}

	level.Debug(logger).Log("msg", "updated token in config file", "tenant", tenant.Tenant)

	return client, nil
}

// Transport returns an OAuth2 HTTP transport based on the current context configuration.
func (c *Config) Transport(ctx context.Context, logger log.Logger) (http.RoundTripper, error) {
	tenant, _, err := c.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("getting current context: %w", err)
	}

	transport, err := tenant.Transport(ctx, logger)
	if err != nil {
		return nil, err
	}

	c.APIs[c.Current.API].Contexts[c.Current.Tenant] = tenant
	if err := c.Save(logger); err != nil {
		return nil, fmt.Errorf("updating token in config file: %w", err)
	}

	level.Debug(logger).Log("msg", "updated token in config file", "tenant", tenant.Tenant)

	return transport, nil
}

// Read loads configuration from disk.
func Read(logger log.Logger) (*Config, error) {
	if err := ensureConfigDir(); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(getConfigFilePath(), os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	cfg := Config{pathOverride: getConfigFilePath()}

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

	file, err := os.OpenFile(getConfigFilePath(), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	level.Debug(logger).Log("msg", "saved config in config file")

	return nil
}

// AddAPI adds a new Observatorium API to the configuration and saves the config to disk.
// In case no name is provided, the hostname of the API URL is used instead.
func (c *Config) AddAPI(logger log.Logger, name string, apiURL string) error {
	if c.APIs == nil {
		c.APIs = make(map[string]APIConfig)
		level.Debug(logger).Log("msg", "initialize config API map")
	}

	url, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL", url)
	}

	// url.Parse might pass a URL with only path, so need to check here for scheme and host.
	// As per docs: https://pkg.go.dev/net/url#Parse.
	if url.Host == "" || url.Scheme == "" {
		return fmt.Errorf("%s is not a valid URL (scheme: %s,host: %s)", url, url.Scheme, url.Host)
	}

	if name == "" {
		// Host name cannot contain slashes, so need not check.
		name = url.Host
		level.Debug(logger).Log("msg", "use hostname as name")
	} else {
		// Need to check due to semantics of context switch.
		if strings.Contains(string(name), "/") {
			return fmt.Errorf("api name %s cannot contain slashes", name)
		}
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
func (c *Config) RemoveAPI(logger log.Logger, name string) error {
	if len(c.APIs) == 1 {
		// Only one API was saved, so can assume it was current context.
		c.APIs = map[string]APIConfig{}
		c.Current.API = ""
		c.Current.Tenant = ""
		level.Debug(logger).Log("msg", "emptied current and removed single config")
		return c.Save(logger)
	}

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
func (c *Config) AddTenant(logger log.Logger, name string, api string, tenant string, oidcCfg *OIDCConfig) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if c.APIs[api].Contexts == nil {
		a := c.APIs[api]
		a.Contexts = make(map[string]TenantConfig)

		c.APIs[api] = a
	}

	// Need to check due to semantics of context switch.
	if strings.Contains(string(name), "/") {
		return fmt.Errorf("tenant name %s cannot contain slashes", name)
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
func (c *Config) RemoveTenant(logger log.Logger, name string, api string) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[name]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", name, api)
	}

	delete(c.APIs[api].Contexts, name)

	return c.Save(logger)
}

func (c *Config) GetContext(api string, tenant string) (TenantConfig, APIConfig, error) {
	if _, ok := c.APIs[api]; !ok {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("api with name %s doesn't exist", c.Current.API)
	}

	if _, ok := c.APIs[api].Contexts[tenant]; !ok {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("tenant with name %s doesn't exist in api %s", c.Current.Tenant, c.Current.API)
	}

	return c.APIs[api].Contexts[tenant], c.APIs[api], nil
}

// GetCurrentContext returns the currently set context i.e, the current API and tenant configuration.
func (c *Config) GetCurrentContext() (TenantConfig, APIConfig, error) {
	if c.Current.API == "" || c.Current.Tenant == "" {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("current context is empty")
	}

	return c.GetContext(c.Current.API, c.Current.Tenant)
}

// SetCurrentContext switches the current context to given api and tenant.
func (c *Config) SetCurrentContext(logger log.Logger, api string, tenant string) error {
	if _, ok := c.APIs[api]; !ok {
		return fmt.Errorf("api with name %s doesn't exist", api)
	}

	if _, ok := c.APIs[api].Contexts[tenant]; !ok {
		return fmt.Errorf("tenant with name %s doesn't exist in api %s", tenant, api)
	}

	if c.Current.API == api && c.Current.Tenant == tenant {
		level.Debug(logger).Log("msg", "context is the same as current")
	}

	c.Current.API = api
	c.Current.Tenant = tenant

	return c.Save(logger)
}

// RemoveContext removes the specified context <api>/<tenant>. If the API configuration has only one tenant,
// the API configuration is removed.
func (c *Config) RemoveContext(logger log.Logger, api string, tenant string) error {
	// If there is only one tenant per API configuration, remove the whole API configuration.
	if _, ok := c.APIs[api].Contexts[tenant]; ok {
		if len(c.APIs[api].Contexts) == 1 {
			return c.RemoveAPI(logger, api)
		}
	}

	return c.RemoveTenant(logger, tenant, api)
}
