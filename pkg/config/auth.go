package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

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

// Client returns an OAuth2 HTTP client based on the current context configuration.
func (c *Config) Client(ctx context.Context, logger log.Logger) (*http.Client, error) {
	tenant, _, err := c.GetCurrent()
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

// TODO(saswatamcode): Replace this with OpenAPI based/dedicated fetcher.
// DoMetricsGetReq makes a GET request to specified endpoint.
func DoMetricsGetReq(ctx context.Context, logger log.Logger, endpoint string) ([]byte, error) {
	config, err := Read(logger)
	if err != nil {
		return nil, fmt.Errorf("getting reading config: %w", err)
	}

	c, err := config.Client(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("getting current client: %w", err)
	}

	resp, err := c.Get(config.APIs[config.Current.API].URL + path.Join("api/metrics/v1", config.APIs[config.Current.API].Contexts[config.Current.Tenant].Tenant, endpoint))
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}

	defer resp.Body.Close()

	level.Debug(logger).Log("msg", "made GET request", "endpoint", endpoint, "status code", resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return b, nil
}
