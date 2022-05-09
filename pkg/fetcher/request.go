package fetcher

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/observatorium/api/client"
	"github.com/observatorium/api/client/parameters"
	"github.com/observatorium/obsctl/pkg/config"
)

// NewCustomFetcher returns a ClientWithResponses which is configured to use oauth HTTP Client.
func NewCustomFetcher(ctx context.Context, logger log.Logger) (*client.ClientWithResponses, parameters.Tenant, error) {
	cfg, err := config.Read(logger)
	if err != nil {
		return nil, "", fmt.Errorf("getting reading config: %w", err)
	}

	c, err := cfg.Client(ctx, logger)
	if err != nil {
		return nil, "", fmt.Errorf("getting current client: %w", err)
	}

	fc, err := client.NewClientWithResponses(cfg.APIs[cfg.Current.API].URL, func(f *client.Client) error {
		f.Client = c
		return nil
	}, client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		level.Debug(logger).Log(
			"method", req.Method,
			"URL", req.URL,
		)
		return nil
	}))
	if err != nil {
		return nil, "", fmt.Errorf("getting fetcher client: %w", err)
	}

	return fc, parameters.Tenant(cfg.Current.Tenant), nil
}
