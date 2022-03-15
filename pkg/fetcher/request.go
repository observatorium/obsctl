package fetcher

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/observatorium/obsctl/pkg/config"
)

// NewCustomFetcher returns a ClientWithResponses which is configured to use oauth HTTP Client.
func NewCustomFetcher(ctx context.Context, logger log.Logger) (*ClientWithResponses, string, error) {
	cfg, err := config.Read(logger)
	if err != nil {
		return nil, "", fmt.Errorf("getting reading config: %w", err)
	}

	c, err := cfg.Client(ctx, logger)
	if err != nil {
		return nil, "", fmt.Errorf("getting current client: %w", err)
	}

	fc, err := NewClientWithResponses(cfg.APIs[cfg.Current.API].URL, func(f *Client) error {
		f.Client = c
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("getting fetcher client: %w", err)
	}

	return fc, cfg.Current.Tenant, nil
}