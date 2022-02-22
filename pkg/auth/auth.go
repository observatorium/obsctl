package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// TODO: Add tests.

const (
	currentTenant = "current_tenant"  // Specifies the key used to identify the currently logged in tenant.
	currentAPI    = "current_api"     // Specifies the key used to identify the current Observatorium API instance.
	configFile    = "obs/config.json" // Path to obsctl auth config file.
)

// getConfigFilePath returns the obsctl config file path or the first string in override arguement.
// Useful for testing.
func getConfigFilePath(override ...string) (string, error) {
	if len(override) != 0 {
		return override[0], nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir path %w", err)
	}

	configFilePath := filepath.Join(configDir, configFile)
	return configFilePath, nil
}

// APIConfig represents configuration for an instance of Observatorium.
type APIConfig struct {
	API  *url.URL
	Name string
}

// TenantConfig represents configuration for a tenant.
type TenantConfig struct {
	APIname string     `json:"apiName"`
	CAFile  []byte     `json:"caFile"`
	Tenant  string     `json:"tenant"`
	OIDC    OIDCConfig `json:"oidc"`
}

// OIDCConfig represents auth configs for a tenant.
type OIDCConfig struct {
	Audience     string `json:"audience"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	IssuerURL    string `json:"issuerURL"`
	Token        *oauth2.Token
}

// savedContext represents the locally saved Tenant and API configuration.
type savedContext struct {
	APIs    map[string]APIConfig
	Tenants map[string]TenantConfig
}

// AddAPI adds a new Observatorium API to saved configuration and sets it to current context. If a name is not provided, it uses URL hostname.
func AddAPI(apiURL string, name string, logger log.Logger) error {
	u, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("parsing API URL %w", err)
	}

	if name == "" {
		name = u.Host
	}

	return APIConfig{Name: name, API: u}.saveAPILocally(logger)
}

// saveAPILocally saves a API configuration in config file. Creates new file if none exists.
func (cfg APIConfig) saveAPILocally(logger log.Logger, override ...string) error {
	configFilePath, err := getConfigFilePath(override...)
	if err != nil {
		return err
	}

	// Config file exists, so update file.
	if _, err := os.Stat(configFilePath); err == nil {
		fileContent, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("reading config file  %w", err)
		}

		var savedContent savedContext
		if err := json.Unmarshal(fileContent, &savedContent); err != nil {
			return fmt.Errorf("unmarshaling saved configs %w", err)
		}

		// Check if API config already exists.
		if v, ok := savedContent.APIs[cfg.Name]; ok {
			// If it exists but isn't the same, then it needs to be updated.
			if !cmp.Equal(v, cfg) {
				level.Info(logger).Log("msg", "API was saved earlier, updated")
			}
		}

		// Handle case where file exists, but no API configs yet.
		if savedContent.APIs == nil {
			savedContent.APIs = make(map[string]APIConfig)
		}

		// Save config and set to current.
		savedContent.APIs[cfg.Name] = cfg
		savedContent.APIs[currentAPI] = cfg

		// Marshal and save to file.
		b, err := json.Marshal(savedContent)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if err := ioutil.WriteFile(configFilePath, b, 0644); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		// File does not exist, so we need to create a new one.
		if err := os.Mkdir(filepath.Dir(configFilePath), os.ModePerm); err != nil {
			return fmt.Errorf("creating config dir %w", err)
		}

		f, err := os.Create(configFilePath)
		if err != nil {
			return fmt.Errorf("creating config file %w", err)
		}
		defer f.Close()

		content := savedContext{}
		content.APIs = make(map[string]APIConfig)
		// Save new API and set it to current.
		content.APIs[cfg.Name] = cfg
		content.APIs[currentAPI] = cfg

		// Marshal and write to file.
		jsonData, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if _, err := f.Write(jsonData); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

		return nil
	} else {
		return err
	}
}

// RemoveAPI removes a locally saved Observatorium API config. Logs if API name not saved.
func RemoveAPI(name string, logger log.Logger, override ...string) error {
	configFilePath, err := getConfigFilePath(override...)
	if err != nil {
		return err
	}

	// Config file exists, so update file.
	if _, err := os.Stat(configFilePath); err == nil {
		fileContent, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("reading config file  %w", err)
		}

		var savedContent savedContext
		if err := json.Unmarshal(fileContent, &savedContent); err != nil {
			return fmt.Errorf("unmarshaling saved configs %w", err)
		}

		// Check if API config exists and remove.
		if _, ok := savedContent.APIs[name]; ok {
			// Also if this is set to current, need to remove that.
			// Fix this (not currently being set).
			if savedContent.APIs[name].Name == savedContent.APIs[currentAPI].Name {
				level.Info(logger).Log("msg", "Removing API that is current context")
				savedContent.APIs[currentAPI] = APIConfig{}
			}
			delete(savedContent.APIs, name)
		} else {
			level.Info(logger).Log("msg", "API with given name does not exist")
			return nil
		}

		// Marshal and save to file.
		b, err := json.Marshal(savedContent)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if err := ioutil.WriteFile(configFilePath, b, 0644); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

	} else if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config file does not exist %w", err)
	} else {
		return err
	}

	return nil
}

// GetCurrentContext returns the currently set context i.e, the current API and tenant configuration.
func GetCurrentContext(override ...string) (TenantConfig, APIConfig, error) {
	configFilePath, err := getConfigFilePath(override...)
	if err != nil {
		return TenantConfig{}, APIConfig{}, err
	}

	// Config file exists, so update file.
	if _, err := os.Stat(configFilePath); err == nil {
		fileContent, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return TenantConfig{}, APIConfig{}, fmt.Errorf("reading config file  %w", err)
		}

		var savedContent savedContext
		if err := json.Unmarshal(fileContent, &savedContent); err != nil {
			return TenantConfig{}, APIConfig{}, fmt.Errorf("unmarshaling saved configs %w", err)
		}

		return savedContent.Tenants[currentTenant], savedContent.APIs[currentAPI], nil
	} else if errors.Is(err, os.ErrNotExist) {
		return TenantConfig{}, APIConfig{}, fmt.Errorf("config file does not exist %w", err)
	} else {
		return TenantConfig{}, APIConfig{}, err
	}
}

// SwitchContext sets the current context to targetContext.
func SwitchContext(targetContext string, logger log.Logger, override ...string) error {
	// Need to split <api name>/<tenant name>.
	s := strings.Split(targetContext, "/")
	if len(s) != 2 {
		return fmt.Errorf("context not correct")
	}

	apiName := s[0]
	tenantName := s[1]

	configFilePath, err := getConfigFilePath(override...)
	if err != nil {
		return err
	}

	// Config file exists, so update file.
	if _, err := os.Stat(configFilePath); err == nil {
		fileContent, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("reading config file  %w", err)
		}

		var savedContent savedContext
		if err := json.Unmarshal(fileContent, &savedContent); err != nil {
			return fmt.Errorf("unmarshaling saved configs %w", err)
		}

		// Check if specified names exist.
		if _, ok := savedContent.Tenants[tenantName]; !ok {
			return fmt.Errorf("specified tenant does not exist")
		}

		if _, ok := savedContent.APIs[apiName]; !ok {
			return fmt.Errorf("specified api does not exist")
		}

		// Warn in case saved tenant API name and given API name is different.
		if savedContent.Tenants[tenantName].APIname != savedContent.APIs[apiName].Name {
			level.Warn(logger).Log("msg", "Selected tenant has different API name set")
		}

		// Set current tenant and API.
		savedContent.Tenants[currentTenant] = savedContent.Tenants[tenantName]
		savedContent.APIs[currentAPI] = savedContent.APIs[apiName]

		// Marshal and save to file.
		b, err := json.Marshal(savedContent)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if err := ioutil.WriteFile(configFilePath, b, 0644); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config file does not exist %w", err)
	} else {
		return err
	}
}

// Login adds the tenant configuration to saved configs and sets it as the current tenant in context.
// Also, requests for token from OIDC provider only if disableOIDCCheck is false.
func Login(ctx context.Context, cfg TenantConfig, disableOIDCCheck bool, logger log.Logger) error {
	return cfg.saveTenantLocally(ctx, disableOIDCCheck, logger)
}

// saveTenantLocally saves a tenant's configuration in config file. Creates new file if none exists.
func (cfg TenantConfig) saveTenantLocally(ctx context.Context, disableOIDCCheck bool, logger log.Logger, override ...string) error {
	configFilePath, err := getConfigFilePath(override...)
	if err != nil {
		return err
	}

	// Config file exists, so update file.
	if _, err := os.Stat(configFilePath); err == nil {
		fileContent, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("reading config file  %w", err)
		}

		var savedContent savedContext
		if err := json.Unmarshal(fileContent, &savedContent); err != nil {
			return fmt.Errorf("unmarshaling saved configs %w", err)
		}

		// Check if tenant config already exists.
		if v, ok := savedContent.Tenants[cfg.Tenant]; ok {
			// If it exists but isn't the same, then it will be updated.
			if !cmp.Equal(v, cfg) {
				level.Info(logger).Log("msg", "Tenant was saved earlier, updated")
			}
		}

		// Handle case where file exists, but no tenant configs yet.
		if savedContent.Tenants == nil {
			savedContent.Tenants = make(map[string]TenantConfig)
		}

		// If OIDC credentials are provided, get token and save.
		if !disableOIDCCheck && !cmp.Equal(cfg.OIDC, OIDCConfig{}) {
			err := cfg.getToken(ctx)
			if err != nil {
				return err
			}
		}

		// Add new entry in map and set current.
		savedContent.Tenants[cfg.Tenant] = cfg
		savedContent.Tenants[currentTenant] = cfg

		// Marshal and save to file.
		b, err := json.Marshal(savedContent)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if err := ioutil.WriteFile(configFilePath, b, 0644); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		// File does not exist, so we need to create a new one.
		if err := os.Mkdir(filepath.Dir(configFilePath), os.ModePerm); err != nil {
			return fmt.Errorf("creating config dir %w", err)
		}

		f, err := os.Create(configFilePath)
		if err != nil {
			return fmt.Errorf("creating config file %w", err)
		}
		defer f.Close()

		content := savedContext{}
		content.Tenants = make(map[string]TenantConfig)

		if !disableOIDCCheck && !cmp.Equal(cfg.OIDC, OIDCConfig{}) {
			err := cfg.getToken(ctx)
			if err != nil {
				return err
			}
		}

		// Save first new tenant and set it to current.
		content.Tenants[cfg.Tenant] = cfg
		content.Tenants[currentTenant] = cfg

		// Marshal and write to file.
		jsonData, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("marshaling configs while saving %w", err)
		}

		if _, err := f.Write(jsonData); err != nil {
			return fmt.Errorf("writing configs to file %w", err)
		}

		return nil
	} else {
		return err
	}
}

// getToken populates Token in cfg.OIDC.
func (cfg *TenantConfig) getToken(ctx context.Context) error {
	if cfg.OIDC.IssuerURL == "" {
		return fmt.Errorf("no issuerURL provided")
	}

	provider, err := oidc.NewProvider(ctx, cfg.OIDC.IssuerURL)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, http.Client{
		Transport: http.DefaultTransport,
	})

	ccc := clientcredentials.Config{
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		TokenURL:     provider.Endpoint().TokenURL,
	}

	if cfg.OIDC.Audience != "" {
		ccc.EndpointParams = url.Values{
			"audience": []string{cfg.OIDC.Audience},
		}
	}

	token, err := ccc.Token(ctx)
	if err != nil {
		return err
	}
	// TODO: use and refresh tokens in fetcher client with auto refresh somehow. Also, incorporate caFile.
	cfg.OIDC.Token = token
	return nil
}
