package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/config"
	"github.com/spf13/cobra"
)

func NewTraceServicesCmd(ctx context.Context) *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "services",
		Short: "List names of services",
		Long:  "List names of services with trace information",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Don't print CLI flag usage if we get network error
			cmd.SilenceUsage = true

			cfg, err := config.Read(logger)
			if err != nil {
				return fmt.Errorf("getting reading config: %w", err)
			}

			client, err := cfg.Client(ctx, logger)
			if err != nil {
				return fmt.Errorf("getting current client: %w", err)
			}

			level.Debug(logger).Log(
				"msg", "Using configuration",
				"URL", cfg.APIs[cfg.Current.API].URL,
				"tenant", cfg.Current.Tenant)

			svcUrl, err := url.Parse(cfg.APIs[cfg.Current.API].URL)
			if err != nil {
				return fmt.Errorf("parsing url: %w", err)
			}
			svcUrl.Path = "api/traces/v1/rhobs/api/services"
			resp, err := client.Get(svcUrl.String())
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			if resp.StatusCode >= 300 {
				level.Debug(logger).Log(
					"msg", "/api/services request failed",
					"statusCode", resp.StatusCode,
					"status", resp.Status,
					"body", string(bodyBytes))
				return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
			}

			switch outputFormat {
			case "table":
				svcs, err := services(bodyBytes)
				if err != nil {
					return fmt.Errorf("parsing services: %w", err)
				}
				if len(svcs) == 0 {
					fmt.Fprintln(cmd.OutOrStderr(), "No services found")
					return nil
				}
				fmt.Fprintln(cmd.OutOrStderr(), "SERVICE")
				for _, svc := range svcs {
					fmt.Fprintln(cmd.OutOrStdout(), svc)
				}
			case "json":
				return prettyPrintJSON(bodyBytes, cmd.OutOrStdout())
			default:
				cmd.SilenceUsage = false
				return fmt.Errorf("unknown format %s", outputFormat)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: json|table")

	return cmd
}

func NewTracesCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Trace-based operations for Observatorium.",
		Long:  "Trace-based operations for Observatorium.",
	}

	cmd.AddCommand(NewTraceServicesCmd(ctx))

	return cmd
}

// services convert the internal Jaeger API /api/services response
// into a list of services
func services(js []byte) ([]string, error) {
	var result map[string]interface{}
	err := json.Unmarshal(js, &result)
	if err != nil {
		return nil, err
	}
	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("no JSON data in %s", string(js))
	}
	if data == nil {
		return []string{}, nil
	}
	services, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected JSON list in %s", string(js))
	}
	retval := make([]string, len(services))
	for i, svc := range services {
		retval[i] = fmt.Sprintf("%s", svc)
	}
	return retval, nil
}
