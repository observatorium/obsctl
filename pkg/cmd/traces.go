package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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

			url := fmt.Sprintf("%s%s", cfg.APIs[cfg.Current.API].URL, "api/traces/v1/rhobs/api/services")
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			if resp.StatusCode >= 300 {
				fmt.Fprintf(os.Stdout, "%d: %s\n%s", resp.StatusCode, resp.Status, string(bodyBytes))
				return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
			}

			if outputFormat == "table" {
				svcs, err := services(bodyBytes)
				if err != nil {
					return fmt.Errorf("parsing services: %w", err)
				}
				_ = printTable(cmd.OutOrStdout(), cmd.OutOrStderr(), svcs)
			} else if outputFormat == "json" {
				return prettyPrintJSON(bodyBytes, cmd.OutOrStdout())
			} else {
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

func printTable(out, err io.Writer, svcs []string) error {
	if len(svcs) == 0 {
		fmt.Fprintln(err, "No services found")
		return nil
	}
	fmt.Fprintln(out, "SERVICE")
	for _, svc := range svcs {
		fmt.Fprintln(out, svc)
	}
	return nil
}
