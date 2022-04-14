package config

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s response: %q", resp.Status, b)
	}

	return b, nil
}

// DoMetricsPutReqWithYAML makes a PUT request to specified endpoint with a YAML body.
func DoMetricsPutReqWithYAML(ctx context.Context, logger log.Logger, endpoint string, body []byte) ([]byte, error) {
	config, err := Read(logger)
	if err != nil {
		return nil, fmt.Errorf("getting reading config: %w", err)
	}

	c, err := config.Client(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("getting current client: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, config.APIs[config.Current.API].URL+path.Join("api/metrics/v1", config.APIs[config.Current.API].Contexts[config.Current.Tenant].Tenant, endpoint), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	req.Header.Set("Content-Type", "application/yaml")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}

	defer resp.Body.Close()

	level.Debug(logger).Log("msg", "made PUT request", "endpoint", endpoint, "status code", resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return b, nil
}
